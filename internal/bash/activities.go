package bash

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kanzihuang/temporal-bash/pkg/bash"
	"go.temporal.io/sdk/temporal"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	prefixSuffixLength = 32 << 10
)

type Activities struct {
	hostTaskQueue string
}

func NewActivities(hostTaskQueue string) *Activities {
	return &Activities{
		hostTaskQueue: hostTaskQueue,
	}
}

// ReadFile read file with temporal, and return error "blob too large" if file size is greater than bash.BlobSizeMax
func (a Activities) ReadFile(_ context.Context, input bash.ReadFileInput) (bash.ReadFileOutput, error) {
	if err := a.matchSessionDir(input.SessionDir); err != nil {
		return bash.ReadFileOutput{}, err
	}
	f, err := os.Open(filepath.Join(input.SessionDir, input.FileName))
	if err != nil {
		return bash.ReadFileOutput{}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	data, err := io.ReadAll(io.LimitReader(f, bash.BlobSizeMax+1))
	if err != nil {
		return bash.ReadFileOutput{}, err
	}
	if len(data) > bash.BlobSizeMax {
		return bash.ReadFileOutput{}, bash.ErrBlobTooLarge
	}
	return bash.ReadFileOutput{
		Data: data,
	}, nil
}

func (a Activities) Begin(_ context.Context, _ bash.BeginInput) (bash.BeginOutput, error) {
	sessionDir, err := os.MkdirTemp(os.TempDir(), a.hostTaskQueue+"-")
	if err != nil {
		return bash.BeginOutput{}, err
	}
	return bash.BeginOutput{
		HostTaskQueue: a.hostTaskQueue,
		SessionDir:    sessionDir,
	}, nil
}

func (a Activities) End(_ context.Context, input bash.EndInput) (bash.EndOutput, error) {
	if err := a.matchSessionDir(input.SessionDir); err != nil {
		return bash.EndOutput{}, err
	}
	if err := os.RemoveAll(input.SessionDir); err != nil {
		return bash.EndOutput{}, err
	}
	return bash.EndOutput{}, nil
}

func (a Activities) matchSessionDir(dir string) error {
	matched, err := filepath.Match(filepath.Join(os.TempDir(), a.hostTaskQueue+"-*"), dir)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("invalid session directory")
	}
	return nil
}

func BuildBash(originCommand string) func(ctx context.Context, input bash.Input) (bash.Output, error) {
	return func(ctx context.Context, input bash.Input) (bash.Output, error) {
		var err error
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		command := os.Expand(originCommand, func(s string) string {
			return "'" + input.Args[s] + "'"
		})
		cmd := exec.CommandContext(ctx, "bash", "-c", command)
		cmd.Stdin = bytes.NewReader(input.StdinData)
		var stdout io.ReadCloser
		if input.WithStdout {
			stdout, err = cmd.StdoutPipe()
			if err != nil {
				return bash.Output{Command: command}, err
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
			return bash.Output{Command: command}, err
		}

		var stdoutData []byte
		if input.WithStdout {
			stdoutData, err = io.ReadAll(io.LimitReader(stdout, bash.BlobSizeMax+1))
			if err != nil {
				return bash.Output{Command: command}, err
			}
			if len(stdoutData) > bash.BlobSizeMax {
				cancel()
				_ = cmd.Wait()
				return bash.Output{Command: command}, fmt.Errorf("stdout data is too large:  %w", bash.ErrBlobTooLarge)
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
			return bash.Output{
				Command:    command,
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, nil
		case errors.As(err, &exitError):
			return bash.Output{
				Command:    command,
				ExitCode:   exitError.ExitCode(),
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, temporal.NewApplicationErrorWithCause()
		default:
			return bash.Output{
				Command:    command,
				ExitCode:   1,
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, err
		}
	}
}
