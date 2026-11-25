package server

import (
	"io"
	"sync"

	"github.com/charmbracelet/ssh"
	"github.com/gdamore/tcell/v2"
)

// sessionTty lets tcell drive an SSH session as its terminal.
//
// tcell's shutdown calls Drain and then waits for its reader to return, but a
// read on an SSH channel only returns when the client sends something. So a
// pump goroutine does the channel reads, and Read waits on the pump or Drain.
// The pump itself ends when the session closes.
type sessionTty struct {
	sess    ssh.Session
	in      chan []byte
	drain   chan struct{}
	once    sync.Once
	pending []byte

	mu     sync.Mutex
	size   tcell.WindowSize
	resize func()
}

func newSessionTty(sess ssh.Session, w ssh.Window) *sessionTty {
	t := &sessionTty{sess: sess, in: make(chan []byte), drain: make(chan struct{}),
		size: tcell.WindowSize{Width: w.Width, Height: w.Height}}
	go t.pump()
	return t
}

func (t *sessionTty) pump() {
	defer close(t.in)
	for {
		buf := make([]byte, 256)
		n, err := t.sess.Read(buf)
		if n > 0 {
			select {
			case t.in <- buf[:n]:
			case <-t.drain:
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func (t *sessionTty) Read(p []byte) (int, error) {
	if len(t.pending) == 0 {
		select {
		case b, ok := <-t.in:
			if !ok {
				return 0, io.EOF
			}
			t.pending = b
		case <-t.drain:
			return 0, io.EOF
		}
	}
	n := copy(p, t.pending)
	t.pending = t.pending[n:]
	return n, nil
}

func (t *sessionTty) Write(p []byte) (int, error) { return t.sess.Write(p) }

func (t *sessionTty) Drain() error {
	t.once.Do(func() { close(t.drain) })
	return nil
}

// Start, Stop and Close are no-ops: the SSH server owns the session's lifetime.
func (t *sessionTty) Start() error { return nil }
func (t *sessionTty) Stop() error  { return nil }
func (t *sessionTty) Close() error { return nil }

func (t *sessionTty) NotifyResize(cb func()) {
	t.mu.Lock()
	t.resize = cb
	t.mu.Unlock()
}

func (t *sessionTty) WindowSize() (tcell.WindowSize, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.size, nil
}

// setSize records a window-change from the client and tells tcell.
func (t *sessionTty) setSize(w ssh.Window) {
	t.mu.Lock()
	t.size = tcell.WindowSize{Width: w.Width, Height: w.Height}
	cb := t.resize
	t.mu.Unlock()
	if cb != nil {
		cb()
	}
}
