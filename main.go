package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ralpioxxcs/gotypes/app"
	"github.com/ralpioxxcs/gotypes/server"
)

func main() {
	addr := flag.String("ssh", "", "serve the game over SSH on this address (e.g. :2222) instead of playing here")
	hostKey := flag.String("hostkey", "gotypes_ed25519", "SSH host key file, created if missing")
	data := flag.String("data", "", "directory for SSH players' settings and history (empty: nothing saved)")
	maxSessions := flag.Int("max-sessions", 50, "maximum concurrent SSH sessions")
	flag.Parse()

	if *addr == "" {
		app.Run()
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := server.ListenAndServe(ctx, server.Config{Addr: *addr, HostKey: *hostKey, DataDir: *data, MaxSessions: *maxSessions})
	if err != nil {
		log.Fatal(err)
	}
}
