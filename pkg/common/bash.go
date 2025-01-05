package common

const (
	Bash = "bash"
)

type BashInput struct {
	Args       map[string]string
	StdinData  []byte
	WithStdout bool
	WithStderr bool
}

type BashOutput struct {
	ExitCode   int
	StdoutData []byte
	StderrData []byte
}
