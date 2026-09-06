package main

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ralscha/sse-eventbus-go"
	"github.com/ralscha/sse-eventbus-go/httpadapter"
)

type dto struct {
	I int    `json:"i"`
	S string `json:"s"`
}

func main() {
	bus, err := sseeventbus.New(
		sseeventbus.WithClientExpiration(time.Minute, time.Minute),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go emitData(ctx, bus)

	server := &http.Server{Addr: ":8080", Handler: newHandler(bus), ReadHeaderTimeout: 5 * time.Second}
	serveErrors := make(chan error, 1)
	go func() {
		log.Printf("backend listening on http://localhost%s", server.Addr)
		serveErrors <- server.ListenAndServe()
	}()

	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-serveErrors:
		stop()
	}

	closeCtx, cancelClose := context.WithTimeout(context.Background(), 5*time.Second)
	if err := bus.Close(closeCtx); err != nil {
		log.Printf("close event bus: %v", err)
	}
	cancelClose()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shut down HTTP server: %v", err)
	}
	cancelShutdown()
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		log.Fatal(serveErr)
	}
}

func newHandler(bus *sseeventbus.Bus) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /register/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := strings.TrimSpace(r.PathValue("id"))
		if clientID == "" || len(clientID) > 128 {
			http.Error(w, "invalid client ID", http.StatusBadRequest)
			return
		}
		if err := httpadapter.Serve(w, r, bus, clientID,
			httpadapter.WithTimeout(0),
			httpadapter.WithWriteTimeout(10*time.Second),
			httpadapter.WithRegistration(sseeventbus.ReplaceSubscriptions(sseeventbus.DefaultEvent, "dto")),
		); err != nil && !errors.Is(err, sseeventbus.ErrClosed) && !errors.Is(err, context.Canceled) {
			log.Printf("SSE client %q: %v", clientID, err)
		}
	})
	if _, err := os.Stat("client/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("client/dist")))
	}

	return mux
}

func emitData(ctx context.Context, bus *sseeventbus.Bus) {
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		values := make([]int, 5)
		for i := range values {
			values[i] = rand.IntN(31)
		}
		if err := bus.Publish(ctx, sseeventbus.NewEvent(values)); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("publish gauges: %v", err)
		}
		if err := bus.Publish(ctx, sseeventbus.NewNamedEventWithData("dto", dto{I: 10, S: "test"})); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("publish dto: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
