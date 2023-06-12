package sub

import (
	"golang.org/x/sys/unix"
	"os/exec"
	"syscall"
)

// ProcGroup is an option to launch a command in a process group.  Provides the capability to send signals to the entire
// process group, as well as convince wrappers around the group.
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

// Signal dispatches the given signal to the process group.  If the process is not running or this has not been attached
// then no signals are attached.
func (p *ProcGroup) Signal(signal syscall.Signal) error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	pgid, err := syscall.Getpgid(p.cmd.Process.Pid)
	if err != nil {
		return err
	}
	return syscall.Kill(-pgid, signal)
}

// Kill sends the SIGKILL to the group
func (p *ProcGroup) Kill() error {
	return p.Signal(unix.SIGKILL)
}

// Terminate sends SIGTERM to the group
func (p *ProcGroup) Terminate() error {
	return p.Signal(unix.SIGTERM)
}

func WithProcGroup() *ProcGroup {
	return &ProcGroup{}
}
