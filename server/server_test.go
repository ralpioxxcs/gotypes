package server

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// start serves on a random local port and returns its address.
func start(t *testing.T, maxSessions int) string {
	t.Helper()
	srv, err := New(Config{HostKey: filepath.Join(t.TempDir(), "hostkey"), MaxSessions: maxSessions})
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(l)
	t.Cleanup(func() { srv.Close() })
	return l.Addr().String()
}

func dial(t *testing.T, addr string) *gossh.Client {
	t.Helper()
	_, key, _ := ed25519.GenerateKey(rand.Reader)
	signer, _ := gossh.NewSignerFromKey(key)
	c, err := gossh.Dial("tcp", addr, &gossh.ClientConfig{
		User:            "player",
		Auth:            []gossh.AuthMethod{gossh.PublicKeys(signer)},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // test server with a throwaway key
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// syncBuf collects session output for polling while the session writes.
type syncBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuf) Write(p []byte) (int, error) { s.mu.Lock(); defer s.mu.Unlock(); return s.b.Write(p) }
func (s *syncBuf) String() string              { s.mu.Lock(); defer s.mu.Unlock(); return s.b.String() }

func waitOutput(t *testing.T, out *syncBuf, want string) {
	t.Helper()
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(20 * time.Millisecond) {
		if strings.Contains(out.String(), want) {
			return
		}
	}
	t.Fatalf("output never contained %q; got %q", want, out.String())
}

// game opens a PTY session running the game and waits for its first screen.
func game(t *testing.T, c *gossh.Client) (*gossh.Session, io.WriteCloser, *syncBuf) {
	t.Helper()
	s, err := c.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	out := &syncBuf{}
	s.Stdout = out
	in, _ := s.StdinPipe()
	if err := s.RequestPty("xterm-256color", 30, 100, gossh.TerminalModes{}); err != nil {
		t.Fatal(err)
	}
	if err := s.Shell(); err != nil {
		t.Fatal(err)
	}
	return s, in, out
}

func TestPlayOverSSH(t *testing.T) {
	c := dial(t, start(t, 5))
	s, in, out := game(t, c)
	waitOutput(t, out, "words")
	if err := s.WindowChange(20, 70); err != nil {
		t.Fatal(err)
	}
	in.Write([]byte("\x03")) // ctrl+c ends the session
	done := make(chan error)
	go func() { done <- s.Wait() }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("session did not end after ctrl+c")
	}
}

func TestRejectsCommandsAndNoPTY(t *testing.T) {
	c := dial(t, start(t, 5))

	s, _ := c.NewSession() // ssh -t host cmd
	s.RequestPty("xterm-256color", 30, 100, gossh.TerminalModes{})
	out, err := s.CombinedOutput("cat /etc/passwd")
	if err == nil || !strings.Contains(string(out), "not supported") {
		t.Fatalf("command with a PTY should be refused: err=%v out=%q", err, out)
	}

	for _, cmd := range []string{"", "cat /etc/passwd"} { // no PTY at all
		s, _ = c.NewSession()
		if out, err := s.CombinedOutput(cmd); err == nil || !strings.Contains(string(out), "PTY") {
			t.Fatalf("%q without a PTY should be refused: err=%v out=%q", cmd, err, out)
		}
	}
}

func TestSessionLimit(t *testing.T) {
	addr := start(t, 1)
	_, _, out := game(t, dial(t, addr))
	waitOutput(t, out, "words")

	s, _ := dial(t, addr).NewSession()
	full := &syncBuf{}
	s.Stdout, s.Stderr = full, full
	s.RequestPty("xterm-256color", 30, 100, gossh.TerminalModes{})
	s.Shell()
	waitOutput(t, full, "server is full")
}

func TestUserDir(t *testing.T) {
	pub := func() gossh.PublicKey {
		p, _, _ := ed25519.GenerateKey(rand.Reader)
		k, _ := gossh.NewPublicKey(p)
		return k
	}
	a, b := pub(), pub()
	if userDir("/d", a) == userDir("/d", b) || userDir("/d", a) != userDir("/d", a) {
		t.Fatal("each key needs its own stable dir")
	}
	if userDir("/d", nil) != "" || userDir("", a) != "" {
		t.Fatal("no key or no data dir must mean no saving")
	}
	if dir := userDir("/d", a); filepath.Dir(dir) != "/d/users" {
		t.Fatalf("unexpected dir %q", dir)
	}
}
