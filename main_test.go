package main

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ralscha/sse-eventbus-go"
)

func TestDTOJSONMatchesJavaRecord(t *testing.T) {
	value, err := json.Marshal(dto{I: 10, S: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != `{"i":10,"s":"test"}` {
		t.Fatalf("unexpected JSON: %s", value)
	}
}

func TestRegisterRejectsBlankClientID(t *testing.T) {
	bus, err := sseeventbus.New(sseeventbus.WithSynchronousDelivery())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = bus.Close(context.Background()) })

	response := httptest.NewRecorder()
	newHandler(bus).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/register/%20", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestRegisterStreamsConvertedEvents(t *testing.T) {
	bus, err := sseeventbus.New(sseeventbus.WithSynchronousDelivery())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = bus.Close(context.Background()) })
	server := httptest.NewServer(newHandler(bus))
	t.Cleanup(server.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/register/client", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = response.Body.Close() }()
	reader := bufio.NewReader(response.Body)
	initial := make([]byte, len(":\n\n"))
	if _, err := io.ReadFull(reader, initial); err != nil || string(initial) != ":\n\n" {
		t.Fatalf("initial SSE frame = %q, err = %v", initial, err)
	}

	deadline := time.Now().Add(time.Second)
	for bus.CountSubscribers(sseeventbus.DefaultEvent) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if bus.CountSubscribers(sseeventbus.DefaultEvent) == 0 {
		t.Fatal("SSE client did not subscribe")
	}
	if err := bus.Publish(ctx, sseeventbus.NewEvent([]int{1, 2})); err != nil {
		t.Fatal(err)
	}
	line, err := reader.ReadString('\n')
	if err != nil || line != "data:[1,2]\n" {
		t.Fatalf("event frame = %q, err = %v", line, err)
	}
}
