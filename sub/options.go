package sub

import "os/exec"

type WorkingDir struct {
	Where string
}

func (w *WorkingDir) Customize(cmd *exec.Cmd) {
	cmd.Dir = w.Where
}
