package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kanzihuang/temporal-shell/pkg/common"
	"io"
	"os"
	"os/exec"
)

const (
	prefixSuffixLength = 32 << 10
)

// ReadFile read file with temporal, and return error "blob too large" if file size is greater than common.BlobSizeMax
func ReadFile(_ context.Context, input common.ReadFileInput) (common.ReadFileOutput, error) {
	f, err := os.Open(input.Path)
	if err != nil {
		return common.ReadFileOutput{}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	data, err := io.ReadAll(io.LimitReader(f, common.BlobSizeMax+1))
	if err != nil {
		return common.ReadFileOutput{}, err
	}
	if len(data) > common.BlobSizeMax {
		return common.ReadFileOutput{}, common.ErrBlobTooLarge
	}
	return common.ReadFileOutput{
		Name: input.Name,
		Path: input.Path,
		Data: data,
	}, nil
}

func BuildGetHostTaskQueue(hostTaskQueue string) func() (string, error) {
	return func() (string, error) {
		return hostTaskQueue, nil
	}
}

func BuildBash(command string) func(ctx context.Context, input common.BashInput) (common.BashOutput, error) {
	return func(ctx context.Context, input common.BashInput) (common.BashOutput, error) {
		var err error
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		command = os.Expand(command, func(s string) string {
			return input.Args[s]
		})
		cmd := exec.CommandContext(ctx, "bash", "-c", command)
		cmd.Stdin = bytes.NewReader(input.StdinData)
		var stdout io.ReadCloser
		if input.WithStdout {
			stdout, err = cmd.StdoutPipe()
			if err != nil {
				return common.BashOutput{}, err
			}
		} else {
			cmd.Stdout = os.Stdout
		}
		if input.WithStderr {
			cmd.Stderr = &prefixSuffixSaver{N: prefixSuffixLength}
		} else {
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Start(); err != nil {
			return common.BashOutput{}, err
		}

		var stdoutData []byte
		if input.WithStdout {
			stdoutData, err = io.ReadAll(io.LimitReader(stdout, common.BlobSizeMax+1))
			if err != nil {
				return common.BashOutput{}, err
			}
			if len(stdoutData) > common.BlobSizeMax {
				cancel()
				_ = cmd.Wait()
				return common.BashOutput{}, fmt.Errorf("stdout data is too large:  %w", common.ErrBlobTooLarge)
			}
		}

		err = cmd.Wait()
		var stderrData []byte
		if input.WithStderr {
			stderrData = cmd.Stderr.(*prefixSuffixSaver).Bytes()
		}

		var exitError *exec.ExitError
		switch {
		case err == nil:
			return common.BashOutput{
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, nil
		case errors.As(err, &exitError):
			return common.BashOutput{
				ExitCode:   exitError.ExitCode(),
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, nil
		default:
			return common.BashOutput{
				ExitCode:   1,
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, err
		}
	}
}
