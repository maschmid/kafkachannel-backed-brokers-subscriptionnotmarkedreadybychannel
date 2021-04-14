package main

import (
	"context"
	"flag"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/types"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var (
	upTo = flag.Int("upTo", 0, "index at which to just ACK the event without returning a new one")

	podNamespace string
	podName      string

	ids uint64

	// Has the final count been received yet?
	finalCount bool
)

// 30kB of data in an event
const size = 30 * 1024

func count(event event.Event) (*event.Event, protocol.Result) {
	count, err := types.ToInteger(event.Context.GetExtensions()["count"])
	if err != nil {
		return nil, protocol.NewReceipt(false, fmt.Sprintf("error evaluating 'count' extension: %v", err))
	}

	if count < (int32)(*upTo) {
		newcount := count + 1
		log.Printf("Got event with count %d, returning %d", count, newcount)

		e := cloudevents.NewEvent()

		id := atomic.AddUint64(&ids, 1)
		e.SetID(fmt.Sprintf("%d", id))
		e.SetSource(fmt.Sprintf("counter/%s/%s", podNamespace, podName))
		e.SetType("countevent")

		e.SetExtension("count", newcount)

		// 30kB
		data := make([]byte, size, size)
		for c := 0; c < size; c++ {
			// anything...
			data[c] = 42
		}

		e.SetData("", data)

		return &e, protocol.ResultACK
	} else {
		log.Printf("acking final count event")

		finalCount = true

		return nil, protocol.ResultACK
	}
}

// Returns "true" if the counter has received the final count event since the last "reset"
func report(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf("%v", finalCount)))
}


// Resets the "finalCount"
func reset(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	finalCount = false
	_, _ = w.Write([]byte("ok\n"))
}

func main() {
	flag.Parse()

	podName = os.Getenv("POD_NAME")
	podNamespace = os.Getenv("POD_NAMESPACE")

	if podName == "" {
		_, _ = fmt.Fprintln(os.Stderr, "No POD_NAME env provided")
		os.Exit(1)
	}

	if podNamespace == "" {
		_, _ = fmt.Fprintln(os.Stderr, "No POD_NAMESPACE env provided")
		os.Exit(1)
	}

	router := http.NewServeMux()
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, count)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}

	router.HandleFunc("/reset", reset)
	router.HandleFunc("/report", report)
	router.Handle("/", h)


	log.Printf("will listen on :8080\n")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("unable to start http server, %s", err)
	}
}
