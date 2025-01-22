package bash

const (
	Begin    string = "Begin"
	End      string = "End"
	ReadFile string = "ReadFile"
)

const (
	// BlobSizeMax Temporal Limit, in MBs, for BLOBs size in an Event when a warning is thrown in the server logs.
	BlobSizeMax = 512 * 1024
)

type ReadFileInput struct {
	SessionDir string
	FileName   string
}

type ReadFileOutput struct {
	Data []byte
}

type BeginInput struct{}
type BeginOutput struct {
	HostTaskQueue string
	SessionDir    string
}

type EndInput struct {
	SessionDir string
}
type EndOutput struct{}
