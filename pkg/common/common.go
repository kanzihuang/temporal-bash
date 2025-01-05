package common

const (
	GetHostTaskQueue string = "GetHostTaskQueue"
	ReadFile         string = "ReadFile"
)

const (
	// BlobSizeMax Temporal Limit, in MBs, for BLOBs size in an Event when a warning is thrown in the server logs.
	BlobSizeMax = 512 * 1024
)

type ReadFileInput struct {
	Name string
	Path string
}

type ReadFileOutput struct {
	Name string
	Path string
	Data []byte
}
