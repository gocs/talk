package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/gocs/talk/talk"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	hub := talk.NewHub()
	go func() {
		defer cancel()
		if err := hub.Run(ctx); err != nil {
			slog.Error("Run", "err", err)
		}
	}()

	http.HandleFunc("/ws/{user}", hub.ServeWS)
	go func() {
		defer cancel()
		slog.InfoContext(ctx, "ListenAndServe", "addr", ":8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Error("ListenAndServe", "err", err)
		}
	}()

	<-ctx.Done()
	slog.Error("Shutting down", "err", ctx.Err())
}
