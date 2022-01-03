package sub

import (
	"io"
	"os/exec"
)

type cmdStdout struct {
	cmd *exec.Cmd
}

func (c *cmdStdout) ObtainSource() (io.ReadCloser, error) {
	return c.cmd.StdoutPipe()
}

func FromCmdStdout(cmd *exec.Cmd) PumpSource {
	return &cmdStdout{cmd: cmd}
}

type cmdStderr struct {
	cmd *exec.Cmd
}

func (c *cmdStderr) ObtainSource() (io.ReadCloser, error) {
	return c.cmd.StderrPipe()
}

func FromCmdStderr(cmd *exec.Cmd) PumpSource {
	return &cmdStderr{cmd: cmd}
}
