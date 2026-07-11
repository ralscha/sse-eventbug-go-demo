package main

import (
	"context"
	"encoding/json"
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
	bus, err := sseeventbus.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go emitData(ctx, bus)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /register/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientID := strings.TrimSpace(r.PathValue("id"))
		w.Header().Set("Cache-Control", "no-store")
		if err := httpadapter.Serve(w, r, bus, clientID,
			httpadapter.WithTimeout(30*time.Second),
			httpadapter.WithRegistration(sseeventbus.SubscribeTo(sseeventbus.DefaultEvent, "dto")),
		); err != nil && !errors.Is(err, sseeventbus.ErrClosed) {
			log.Printf("SSE client %q: %v", clientID, err)
		}
	})
	if _, err := os.Stat("client/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("client/dist")))
	}

	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("backend listening on http://localhost%s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	_ = bus.Close(shutdownCtx)
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
		encoded, _ := json.Marshal(values)
		if err := bus.Publish(ctx, sseeventbus.NewEvent(string(encoded))); err != nil && !errors.Is(err, context.Canceled) {
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
