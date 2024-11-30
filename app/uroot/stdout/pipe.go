package stdout

import (
	"bufio"
	"context"
	"io"
	"os"
)

func PipeStdOut(ctx context.Context, from io.ReadCloser) {
	pipe(ctx, from, os.Stdout)
}

func PipeStdErr(ctx context.Context, from io.ReadCloser) {
	pipe(ctx, from, os.Stderr)
}

func pipe(ctx context.Context, from io.ReadCloser, to *os.File) {
	newCtx, _ := context.WithCancel(ctx)
	go func() {
		scanner := bufio.NewScanner(from)
		for scanner.Scan() {
			if scanner.Err() != nil {
				break
			}

			select {
			case <-newCtx.Done():
				from.Close()
				break
			default:
				to.WriteString(scanner.Text()) // Print to parent's stdout
				to.WriteString("\n")
			}
		}
	}()
}
