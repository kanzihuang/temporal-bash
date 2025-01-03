package shell

import (
	"github.com/kanzihuang/temporal-shell/pkg/common"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"os"
	"path/filepath"
	"testing"
)

const hostTaskQueue = "testHostTaskQueue"

func TestActivityTestSuite(t *testing.T) {
	suite.Run(t, new(ActivityTestSuite))
}

type ActivityTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestActivityEnvironment
}

func (s *ActivityTestSuite) SetupTest() {
	s.env = s.NewTestActivityEnvironment()
	s.env.RegisterActivityWithOptions(BuildGetHostTaskQueue(hostTaskQueue), activity.RegisterOptions{Name: common.GetHostTaskQueue})
	s.env.RegisterActivityWithOptions(ReadFile, activity.RegisterOptions{Name: common.ReadFile})
}

func (s *ActivityTestSuite) TestGetHostTaskQueue() {
	val, err := s.env.ExecuteActivity(common.GetHostTaskQueue)
	s.NoError(err)
	s.True(val.HasValue())

	var taskQueue string
	err = val.Get(&taskQueue)
	s.NoError(err)
	s.Equal(hostTaskQueue, taskQueue)
}

func (s *ActivityTestSuite) beforeTestReadFile(path string, data []byte) {
	err := os.WriteFile(path, data, 0666)
	s.NoError(err)
}

func (s *ActivityTestSuite) afterTestReadFile(path string) {
	err := os.Remove(path)
	s.NoError(err)
}

func (s *ActivityTestSuite) TestReadFile() {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "test-read-file-ok",
			data:    []byte("hello world"),
			wantErr: false,
		},
		{
			name:    "test-read-file-valid-size",
			data:    make([]byte, common.BlobSizeMax),
			wantErr: false,
		},
		{
			name:    "test-read-file-too-large",
			data:    make([]byte, common.BlobSizeMax+1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := filepath.Join(os.TempDir(), tt.name)
			s.beforeTestReadFile(path, tt.data)
			defer s.afterTestReadFile(path)

			require := s.Require()
			val, err := s.env.ExecuteActivity(common.ReadFile, common.ReadFileInput{
				Name: tt.name,
				Path: path,
			})
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.True(val.HasValue())

			var output common.ReadFileOutput
			err = val.Get(&output)
			require.NoError(err)
			require.Equal(tt.name, output.Name)
			require.Equal(path, output.Path)
			require.Equal(tt.data, output.Data)
		})
	}
}
