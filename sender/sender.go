package main

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var (
	sink         string
	podNamespace string
	podName      string

	ids uint64
)

// 30kB of data in an event
const size = 30 * 1024

func send() error {
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		return err
	}

	ctx := cloudevents.ContextWithTarget(context.Background(), sink)
	e := cloudevents.NewEvent()

	id := atomic.AddUint64(&ids, 1)
	e.SetID(fmt.Sprintf("%d", id))
	e.SetSource(fmt.Sprintf("sender/%s/%s", podNamespace, podName))
	e.SetType("countevent")

	e.SetExtension("count", 0)

	// 30kB
	data := make([]byte, size, size)
	for c := 0; c < size; c++ {
		// anything...
		data[c] = 42
	}

	e.SetData("", data)

	result := c.Send(ctx, e)
	if !cloudevents.IsACK(result) {
		s := fmt.Sprintf("error sending event: %v", result)
		log.Println(s)
	}

	return nil
}

func main() {
	sink = os.Getenv("K_SINK")
	podName = os.Getenv("POD_NAME")
	podNamespace = os.Getenv("POD_NAMESPACE")

	if sink == "" {
		_, _ = fmt.Fprintln(os.Stderr, "No K_SINK env provided")
		os.Exit(1)
	}

	if podName == "" {
		_, _ = fmt.Fprintln(os.Stderr, "No POD_NAME env provided")
		os.Exit(1)
	}

	if podNamespace == "" {
		_, _ = fmt.Fprintln(os.Stderr, "No POD_NAMESPACE env provided")
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := send()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
