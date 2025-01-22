package bash

type Input struct {
	Args       map[string]string
	StdinData  []byte
	WithStdout bool
	WithStderr bool
}

type Output struct {
	Command    string
	ExitCode   int
	StdoutData []byte
	StderrData []byte
}
