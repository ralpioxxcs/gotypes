// Package server serves gotypes over SSH: every session gets its own game.
package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"path/filepath"
	"slices"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/logging"
	"github.com/charmbracelet/wish/ratelimiter"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	"github.com/ralpioxxcs/gotypes/app"
	"github.com/ralpioxxcs/gotypes/internal/race"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/time/rate"
)

type Config struct {
	Addr        string // e.g. ":2222"
	HostKey     string // path; created on first start, keep it so clients don't see a new key
	DataDir     string // per-key settings and history; "" = nothing saved
	MaxSessions int
}

// New builds the SSH server. Anyone may connect: public-key users get saved
// settings and history (per key), everyone else plays without saving.
func New(cfg Config) (*ssh.Server, error) {
	return wish.NewServer(
		wish.WithAddress(cfg.Addr),
		wish.WithHostKeyPath(cfg.HostKey),
		wish.WithIdleTimeout(10*time.Minute),
		wish.WithMaxTimeout(2*time.Hour),
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		wish.WithKeyboardInteractiveAuth(func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool { return true }),
		// listed innermost first: logging runs first, play last
		wish.WithMiddleware(
			play(cfg.DataDir, race.NewHub(app.RaceText, time.Now)),
			activeterm.Middleware(), // no PTY, no game
			limit(cfg.MaxSessions),
			ratelimiter.Middleware(ratelimiter.NewRateLimiter(rate.Every(time.Second), 10, 1000)), // per IP
			logging.Middleware(),
		),
	)
}

// ListenAndServe runs the server until ctx ends, then closes open sessions.
func ListenAndServe(ctx context.Context, cfg Config) error {
	srv, err := New(cfg)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()
	log.Printf("gotypes: ssh server on %s", cfg.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, ssh.ErrServerClosed) {
		return err
	}
	return nil
}

func limit(max int) wish.Middleware {
	slots := make(chan struct{}, max)
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
				next(s)
			default:
				wish.Fatalln(s, "gotypes: the server is full, try again later")
			}
		}
	}
}

func play(dataDir string, hub *race.Hub) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			if len(s.Command()) > 0 {
				wish.Fatalln(s, "gotypes: commands are not supported, just ssh in")
				return
			}
			pty, winCh, _ := s.Pty() // activeterm guarantees a PTY
			tty := newSessionTty(s, pty.Window)
			go func() { // ends when the session closes winCh
				for w := range winCh {
					tty.setSize(w)
				}
			}()
			scr, err := tcell.NewTerminfoScreenFromTtyTerminfo(tty, terminfoFor(pty.Term, s.Environ()))
			if err == nil {
				err = app.Play(s.Context(), scr, userDir(dataDir, s.PublicKey()), hub, s.User())
			}
			if err != nil {
				log.Printf("gotypes: session %s: %v", s.RemoteAddr(), err)
			}
			next(s)
		}
	}
}

// terminfoFor returns a private copy of the client's terminfo (tcell's lookup
// shares one per name), with 24-bit color on unless the client sends
// TCELL_TRUECOLOR=disable (ssh -o SetEnv=TCELL_TRUECOLOR=disable).
func terminfoFor(term string, env []string) *terminfo.Terminfo {
	ti, err := tcell.LookupTerminfo(term)
	if err != nil {
		ti, _ = tcell.LookupTerminfo("xterm-256color") // built into tcell
	}
	cp := *ti
	if slices.Contains(env, "TCELL_TRUECOLOR=disable") {
		cp.SetFgRGB, cp.SetBgRGB, cp.SetFgBgRGB = "", "", ""
	} else if cp.SetFgBgRGB == "" {
		cp.SetFgRGB = "\x1b[38;2;%p1%d;%p2%d;%p3%dm"
		cp.SetBgRGB = "\x1b[48;2;%p1%d;%p2%d;%p3%dm"
		cp.SetFgBgRGB = "\x1b[38;2;%p1%d;%p2%d;%p3%d;48;2;%p4%d;%p5%d;%p6%dm"
	}
	return &cp
}

// userDir is where a public key's data lives, named by a hash of the key so
// nothing from the client ends up in a path.
func userDir(dataDir string, key ssh.PublicKey) string {
	if dataDir == "" || key == nil {
		return ""
	}
	sum := sha256.Sum256(key.Marshal())
	return filepath.Join(dataDir, "users", hex.EncodeToString(sum[:16]))
}
