package shell

import (
	"context"
	"github.com/kanzihuang/temporal-shell/pkg/common"
	"io"
	"os"
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

func BuildExecute(command string) func(ctx context.Context, input common.ActivityInput) (common.ActivityOutput, error) {
	return func(ctx context.Context, input common.ActivityInput) (common.ActivityOutput, error) {
		panic("implement me")
	}
}
