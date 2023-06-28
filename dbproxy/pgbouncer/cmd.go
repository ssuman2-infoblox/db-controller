package pgbouncer

import (
	"io"
	"os/exec"
	"syscall"
)

// execCmd is an interface for abstracting exec.Cmd.
type execCmd interface {
	Start() error
	Wait() error
	Kill() error
	Sighup() error
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
}

// mockableCmd is a mockable exec.Cmd wrapper.
type mockableCmd struct {
	cmd *exec.Cmd
}

func (c *mockableCmd) Start() error {
	return c.cmd.Start()
}
func (c *mockableCmd) Wait() error {
	return c.cmd.Wait()
}
func (c *mockableCmd) Kill() error {
	return c.cmd.Process.Signal(syscall.SIGTERM)
}

func (c *mockableCmd) SetStdout(stdout io.Writer) {
	c.cmd.Stdout = stdout
}

func (c *mockableCmd) SetStderr(stderr io.Writer) {
	c.cmd.Stderr = stderr
}

func (c *mockableCmd) Sighup() error {
	return c.cmd.Process.Signal(syscall.SIGHUP)
}

var _ execCmd = &mockableCmd{}

var (
	newExecCommand = func(name string, args ...string) *mockableCmd {

		mockableCmd := &mockableCmd{
			cmd: exec.Command(name, args...),
		}
		return mockableCmd
	}

	execLookPath func(file string) (string, error) = exec.LookPath
)
