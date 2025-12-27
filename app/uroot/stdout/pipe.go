package stdout

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/discue/go-syscall-gatekeeper/app/utils"
)

func PipeStdOut(ctx context.Context, from io.ReadCloser) {
	pipe(ctx, from, os.Stdout)
}

func PipeStdErr(ctx context.Context, from io.ReadCloser) {
	pipe(ctx, from, os.Stderr)
}

func pipe(ctx context.Context, from io.ReadCloser, to *os.File) {
	go func() {
		scanner := bufio.NewScanner(from)
	forLoop:
		for scanner.Scan() {
			if scanner.Err() != nil {
				break forLoop
			}

			select {
			case <-ctx.Done():
				utils.SafeClose(from, "stdout/stderr pipe reader")
				break forLoop
			default:
				_, _ = to.WriteString(scanner.Text()) // Print to parent's stdout
				_, _ = to.WriteString("\n")
			}
		}
	}()
}
