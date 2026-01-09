package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type testCase struct {
	Describe      string   `yaml:"describe"`
	ExpectFailure bool     `yaml:"expect_failure"`
	Permissions   []string `yaml:"permissions"`
	// When true, do not wrap the container with /gatekeeper. Instead,
	// pass permissions to the container and let its entrypoint handle
	// starting gate-kept processes (e.g., the server) while running
	// other tools (e.g., curl) separately.
	// Defaults to true when omitted in YAML.
	ConfigureGatekeeperEntrypoint *bool `yaml:"configure_gatekeeper_entrypoint"`
}

type suite struct {
	Dir            string
	YamlPath       string
	DockerfilePath string
	CmdArgs        []string
	Cases          []testCase
	ImageTag       string
}

func main() {
	var (
		failFast    bool
		onlyPattern string
		timeoutSec  int
	)

	flag.BoolVar(&failFast, "fail-fast", false, "stop on first failure")
	flag.StringVar(&onlyPattern, "only", "", "only run suites matching this regex pattern (applied to tests.yml path)")
	flag.IntVar(&timeoutSec, "timeout-sec", 20, "per-test timeout in seconds")
	flag.Parse()

	repoRoot, err := findRepoRoot()
	must(err)

	fmt.Println("Building gatekeeper binary locally…")
	gatekeeperBin, cleanup, err := buildGatekeeperLocally(repoRoot)
	must(err)
	defer cleanup()

	testRoot := filepath.Join(repoRoot, "test-e2e")
	suites, err := discoverSuites(testRoot, onlyPattern)
	must(err)
	if len(suites) == 0 {
		fmt.Println("No tests.yml found; nothing to run.")
		os.Exit(0)
	}

	fmt.Printf("Found %d suite(s). Building images…\n", len(suites))
	for i := range suites {
		s := &suites[i]

		// Prepare context: copy gatekeeper binary next to Dockerfile as 'gatekeeper'.
		stagedBin := filepath.Join(s.Dir, "gatekeeper")
		if err := copyFile(gatekeeperBin, stagedBin, 0o755); err != nil {
			must(fmt.Errorf("stage gatekeeper into %s: %w", s.Dir, err))
		}

		// Build image
		s.ImageTag = imageTagFromPath(s.Dir)
		err := dockerBuild(s.ImageTag, s.Dir)

		// Remove staged binary after image build; the image contains /gatekeeper now
		_ = os.Remove(stagedBin)
		if err != nil {
			must(fmt.Errorf("docker build failed for %s: %w", s.Dir, err))
		}

		// Parse CMD once per suite
		cmdArgs, err := parseDockerfileCMD(s.DockerfilePath)
		must(err)
		s.CmdArgs = cmdArgs
	}

	fmt.Println("Running tests…")
	var failures []string
	for _, s := range suites {
		for idx, tc := range s.Cases {
			label := fmt.Sprintf("%s [%d] %s", relPath(testRoot, s.YamlPath), idx+1, tc.Describe)
			fmt.Printf("\x1b[34mpending\x1b[0m %s\r", label)
			ok, timedOut, out := runOne(testRoot, &s, tc, time.Duration(timeoutSec)*time.Second)
			// Clear the pending line
			fmt.Print("\r\x1b[K")
			if ok {
				fmt.Printf("\x1b[32mok\x1b[0m %s\n", label)
			} else if timedOut {
				fmt.Printf("timeout %s\n", label)
				failures = append(failures, fmt.Sprintf("timeout -> %s", s.YamlPath))
				if failFast {
					goto END
				}
			} else {
				fmt.Printf("\x1b[91mfailed\x1b[0m %s\n", label)
				if out != "" {
					fmt.Println(out)
				}
				failures = append(failures, fmt.Sprintf("failed -> %s", s.YamlPath))
				if failFast {
					goto END
				}
			}
		}
	}

END:
	if len(failures) > 0 {
		fmt.Printf("\nYou have %d test failure(s):\n", len(failures))
		for _, f := range failures {
			fmt.Println(f)
		}
		os.Exit(1)
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fi, err := os.Stat(filepath.Join(dir, "main.go")); err == nil && !fi.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached volume root
			return "", errors.New("could not locate repo root (main.go)")
		}
		dir = parent
	}
}

func buildGatekeeperLocally(repoRoot string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "gatekeeper-bin-")
	if err != nil {
		return "", func() {}, err
	}
	outBin := filepath.Join(tmpDir, "gatekeeper")

	// Default to linux/amd64 to run inside Debian/Ubuntu containers.
	env := os.Environ()
	if os.Getenv("GOOS") == "" {
		env = append(env, "GOOS=linux")
	}
	if os.Getenv("GOARCH") == "" {
		env = append(env, "GOARCH=amd64")
	}
	if os.Getenv("CGO_ENABLED") == "" {
		env = append(env, "CGO_ENABLED=0")
	}

	cmd := exec.Command("go", "build", "-o", outBin, filepath.Join(repoRoot, "main.go"))
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Fallback: try host build (may not run in Linux containers if host is non-Linux)
		fmt.Fprintln(os.Stderr, "cross-compile failed, attempting host build (may not run in Linux containers)…")
		cmd2 := exec.Command("go", "build", "-o", outBin, filepath.Join(repoRoot, "main.go"))
		cmd2.Stdout = os.Stdout
		cmd2.Stderr = os.Stderr
		if err2 := cmd2.Run(); err2 != nil {
			return "", func() { _ = os.RemoveAll(tmpDir) }, fmt.Errorf("local build failed: %w (fallback: %v)", err, err2)
		}
	}
	_ = os.Chmod(outBin, 0o755)
	return outBin, func() { _ = os.RemoveAll(tmpDir) }, nil
}

func discoverSuites(testRoot, onlyPattern string) ([]suite, error) {
	var suites []suite
	var re *regexp.Regexp
	var err error
	if onlyPattern != "" {
		re, err = regexp.Compile(onlyPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid --only pattern: %w", err)
		}
	}

	err = filepath.WalkDir(testRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "tests.yml") {
			if re != nil && !re.MatchString(path) {
				return nil
			}

			dir := filepath.Dir(path)
			dockerfile := filepath.Join(dir, "Dockerfile")
			if _, err := os.Stat(dockerfile); err != nil {
				// Skip suites without Dockerfile
				return nil
			}
			cases, err := readTestCases(path)
			if err != nil {
				return err
			}
			suites = append(suites, suite{Dir: dir, YamlPath: path, DockerfilePath: dockerfile, Cases: cases})
		}
		return nil
	})
	return suites, err
}

func readTestCases(path string) ([]testCase, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list []testCase
	if err := yaml.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func parseDockerfileCMD(dockerfile string) ([]string, error) {
	f, err := os.Open(dockerfile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	// scan for last CMD occurrence
	var line string
	for scanner.Scan() {
		txt := strings.TrimSpace(scanner.Text())
		if len(txt) == 0 || strings.HasPrefix(strings.ToUpper(txt), "#") {
			continue
		}
		up := strings.ToUpper(txt)
		if strings.HasPrefix(up, "CMD ") || (len(up) >= 3 && up[:3] == "CMD") {
			line = txt
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if line == "" {
		return nil, fmt.Errorf("no CMD found in %s", dockerfile)
	}

	// Try JSON-array form: CMD ["bin","arg"]
	if i := strings.Index(line, "["); i >= 0 {
		var arr []string
		jsonPart := line[i:]
		if err := json.Unmarshal([]byte(jsonPart), &arr); err != nil {
			return nil, fmt.Errorf("parse CMD array: %w", err)
		}
		return arr, nil
	}
	// Shell form: CMD some command string
	parts := strings.Fields(strings.TrimPrefix(line, "CMD "))
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty CMD in %s", dockerfile)
	}
	return parts, nil
}

func imageTagFromPath(dir string) string {
	rel := strings.ReplaceAll(dir, string(os.PathSeparator), "/")
	sum := sha1.Sum([]byte(rel))
	return fmt.Sprintf("gatekeeper-e2e-%x", sum[:6])
}

func dockerBuild(tag, contextDir string) error {
	cmd := exec.Command("docker", "build", "-t", tag, contextDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runOne(testRoot string, s *suite, tc testCase, timeout time.Duration) (ok bool, timedOut bool, out string) {
	var args []string
	// Default to true if not specified
	cfgEntrypoint := true
	if tc.ConfigureGatekeeperEntrypoint != nil {
		cfgEntrypoint = *tc.ConfigureGatekeeperEntrypoint
	}

	if cfgEntrypoint {
		// Default: wrap whole container with gatekeeper.
		// Compose args: /gatekeeper trace <permissions> -- <CMD from Dockerfile>
		args = []string{"run", "--rm", "--entrypoint", "/gatekeeper", s.ImageTag, "run"}
		if len(tc.Permissions) > 0 {
			args = append(args, tc.Permissions...)
		}
		args = append(args, "--")
		args = append(args, s.CmdArgs...)
	} else {

		// Let container's CMD/entrypoint run as-is, and pass permissions
		// via environment for the script to consume.
		args = []string{"run", "--rm"}
		if len(tc.Permissions) > 0 {
			// Space-join flags for easy parsing in shell
			joined := strings.Join(tc.Permissions, " ")
			args = append(args, "-e", "SERVER_PERMISSIONS="+joined)
		}
		args = append(args, s.ImageTag)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)
	var b strings.Builder
	cmd.Stdout = &b
	cmd.Stderr = &b
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		// Best effort force cleanup: nothing to do since --rm
		return false, true, b.String()
	}
	// Success criteria depends on expectation
	exitCode := exitStatus(err)
	if tc.ExpectFailure {
		return exitCode != 0, false, b.String()
	}
	return exitCode == 0, false, b.String()
}

func exitStatus(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if status, ok := ee.Sys().(interface{ ExitStatus() int }); ok {
			return status.ExitStatus()
		}
	}
	return 1
}

func relPath(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		return p
	}
	return r
}

func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if mode != 0 {
		_ = os.Chmod(dst, mode)
	}
	return nil
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
