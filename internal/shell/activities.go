package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kanzihuang/temporal-shell/pkg/shell"
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

// ReadFile read file with temporal, and return error "blob too large" if file size is greater than shell.BlobSizeMax
func (a Activities) ReadFile(_ context.Context, input shell.ReadFileInput) (shell.ReadFileOutput, error) {
	if err := a.matchSessionDir(input.SessionDir); err != nil {
		return shell.ReadFileOutput{}, err
	}
	f, err := os.Open(filepath.Join(input.SessionDir, input.FileName))
	if err != nil {
		return shell.ReadFileOutput{}, err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	data, err := io.ReadAll(io.LimitReader(f, shell.BlobSizeMax+1))
	if err != nil {
		return shell.ReadFileOutput{}, err
	}
	if len(data) > shell.BlobSizeMax {
		return shell.ReadFileOutput{}, shell.ErrBlobTooLarge
	}
	return shell.ReadFileOutput{
		Data: data,
	}, nil
}

func (a Activities) Begin(_ context.Context, _ shell.BeginInput) (shell.BeginOutput, error) {
	sessionDir, err := os.MkdirTemp(os.TempDir(), a.hostTaskQueue+"-")
	if err != nil {
		return shell.BeginOutput{}, err
	}
	return shell.BeginOutput{
		HostTaskQueue: a.hostTaskQueue,
		SessionDir:    sessionDir,
	}, nil
}

func (a Activities) End(_ context.Context, input shell.EndInput) (shell.EndOutput, error) {
	if err := a.matchSessionDir(input.SessionDir); err != nil {
		return shell.EndOutput{}, err
	}
	if err := os.RemoveAll(input.SessionDir); err != nil {
		return shell.EndOutput{}, err
	}
	return shell.EndOutput{}, nil
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

func BuildBash(originCommand string) func(ctx context.Context, input shell.BashInput) (shell.BashOutput, error) {
	return func(ctx context.Context, input shell.BashInput) (shell.BashOutput, error) {
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
				return shell.BashOutput{Command: command}, err
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
			return shell.BashOutput{Command: command}, err
		}

		var stdoutData []byte
		if input.WithStdout {
			stdoutData, err = io.ReadAll(io.LimitReader(stdout, shell.BlobSizeMax+1))
			if err != nil {
				return shell.BashOutput{Command: command}, err
			}
			if len(stdoutData) > shell.BlobSizeMax {
				cancel()
				_ = cmd.Wait()
				return shell.BashOutput{Command: command}, fmt.Errorf("stdout data is too large:  %w", shell.ErrBlobTooLarge)
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
			return shell.BashOutput{
				Command:    command,
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, nil
		case errors.As(err, &exitError):
			return shell.BashOutput{
				Command:    command,
				ExitCode:   exitError.ExitCode(),
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, nil
		default:
			return shell.BashOutput{
				Command:    command,
				ExitCode:   1,
				StdoutData: stdoutData,
				StderrData: stderrData,
			}, err
		}
	}
}
