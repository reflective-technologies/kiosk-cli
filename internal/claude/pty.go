package claude

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/muesli/cancelreader"
	"golang.org/x/term"
)

// ErrDetached indicates the user detached from the session.
var ErrDetached = errors.New("session detached")

const (
	DefaultDetachKey        = 0x0b // ctrl+k
	DefaultInterruptDelay   = 50 * time.Millisecond
	DefaultInterruptTimeout = 3 * time.Second
)

// SessionIO provides input/output handles for a session.
type SessionIO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// SessionOptions configures PTY execution behavior.
type SessionOptions struct {
	IO               SessionIO
	DetachKey        byte
	InterruptDelay   time.Duration
	InterruptTimeout time.Duration
}

// RunWithPTY starts the command under a PTY, proxies IO, and supports detach.
func RunWithPTY(cmd *exec.Cmd, opts SessionOptions) error {
	if cmd == nil {
		return errors.New("nil command")
	}

	ioCfg := opts.IO
	if ioCfg.Stdin == nil {
		ioCfg.Stdin = os.Stdin
	}
	if ioCfg.Stdout == nil {
		ioCfg.Stdout = os.Stdout
	}
	if ioCfg.Stderr == nil {
		ioCfg.Stderr = os.Stderr
	}

	detachKey := opts.DetachKey
	if detachKey == 0 {
		detachKey = DefaultDetachKey
	}

	interruptDelay := opts.InterruptDelay
	if interruptDelay == 0 {
		interruptDelay = DefaultInterruptDelay
	}

	interruptTimeout := opts.InterruptTimeout
	if interruptTimeout == 0 {
		interruptTimeout = DefaultInterruptTimeout
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer ptmx.Close()

	restoreFn, err := makeRawIfPossible(ioCfg.Stdin)
	if err == nil && restoreFn != nil {
		defer restoreFn()
	}

	resizeCh := make(chan os.Signal, 1)
	defer close(resizeCh)
	if tty, ok := ioCfg.Stdout.(*os.File); ok {
		_ = pty.InheritSize(ptmx, tty)
		signal.Notify(resizeCh, syscall.SIGWINCH)
		go func() {
			for range resizeCh {
				_ = pty.InheritSize(ptmx, tty)
			}
		}()
	}
	defer signal.Stop(resizeCh)

	go func() {
		_, _ = io.Copy(ioCfg.Stdout, ptmx)
	}()

	cr, err := cancelreader.NewReader(ioCfg.Stdin)
	if err != nil {
		return err
	}
	defer cr.Cancel()

	inputErr := make(chan error, 1)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := cr.Read(buf)
			if n > 0 {
				for i := 0; i < n; i++ {
					if buf[i] == detachKey {
						inputErr <- ErrDetached
						return
					}
					if _, werr := ptmx.Write(buf[i : i+1]); werr != nil {
						inputErr <- werr
						return
					}
				}
			}
			if err != nil {
				inputErr <- err
				return
			}
		}
	}()

	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
	}()

	for {
		select {
		case err := <-waitErr:
			return err
		case err := <-inputErr:
			if errors.Is(err, ErrDetached) {
				return detach(cmd, waitErr, interruptDelay, interruptTimeout)
			}
			if errors.Is(err, cancelreader.ErrCanceled) {
				continue
			}
			return err
		}
	}
}

func detach(cmd *exec.Cmd, waitErr <-chan error, delay, timeout time.Duration) error {
	sendInterrupts(cmd.Process, delay)

	select {
	case <-waitErr:
		return ErrDetached
	case <-time.After(timeout):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-waitErr
		return ErrDetached
	}
}

func sendInterrupts(proc *os.Process, delay time.Duration) {
	if proc == nil {
		return
	}
	sendInterrupt(proc)
	time.Sleep(delay)
	sendInterrupt(proc)
}

func sendInterrupt(proc *os.Process) {
	if proc == nil {
		return
	}
	pid := proc.Pid
	if pid > 0 {
		_ = syscall.Kill(-pid, syscall.SIGINT)
	}
	_ = proc.Signal(os.Interrupt)
}

func makeRawIfPossible(r io.Reader) (func() error, error) {
	f, ok := r.(interface{ Fd() uintptr })
	if !ok {
		return nil, nil
	}

	fd := int(f.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	return func() error {
		return term.Restore(fd, oldState)
	}, nil
}
