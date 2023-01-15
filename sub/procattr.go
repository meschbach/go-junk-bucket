package sub

import (
	"golang.org/x/sys/unix"
	"os/exec"
	"syscall"
)

type ProcGroup struct {
	cmd *exec.Cmd
}

func (p *ProcGroup) Customize(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setpgid = true
	p.cmd = cmd
}

func (p *ProcGroup) Signal(signal syscall.Signal) error {
	pgid, err := syscall.Getpgid(p.cmd.Process.Pid)
	if err != nil {
		return err
	}
	return syscall.Kill(-pgid, signal)
}

func (p *ProcGroup) Kill() error {
	return p.Signal(unix.SIGKILL)
}

func (p *ProcGroup) Terminate() error {
	return p.Signal(unix.SIGTERM)
}

func WithProcGroup() *ProcGroup {
	return &ProcGroup{}
}
