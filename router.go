package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	readerctx "github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/event"
	_ "github.com/bcurnow/magicband-reader/handler"
)

type Router interface {
	Route(event event.Event) error
	Close()
	Closed() bool
}

type router struct {
	server            *http.Server
	webChannel        chan event.Event
	webRequestChannel chan bool
	closed            bool
	listenAddress     string
	listenPort        int
}

func NewRouter(listenAddress string, listenPort int) (*router, error) {
	router := router{listenAddress: listenAddress, listenPort: listenPort}
	router.init()
	return &router, nil
}

func (r *router) Route(event event.Event) error {
	log.Tracef("Starting Route, state: %#v", readerctx.State)
	select {
	case <-r.webRequestChannel:
		// There is a web request waiting for an event
		r.webChannel <- event
		log.Tracef("Web Route complete, state: %#v", readerctx.State)
	default:
		// There is no web request waiting, send to handlers
		r.handle(event)
		log.Tracef("Default Route complete, state: %#v", readerctx.State)
	}
	return nil
}

func (r *router) Close() {
	log.Trace("Closing Router")
	r.closed = true
	log.Trace("Shutting down the server")
	if err := r.server.Shutdown(context.Background()); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}
	log.Trace("Server shutdown")
	close(r.webChannel)
	close(r.webRequestChannel)
	log.Trace("Router closed")
}

func (r *router) Closed() bool {
	return r.closed
}

func (r *router) init() error {
	r.server = r.createServer()
	r.webChannel = make(chan event.Event)
	r.webRequestChannel = make(chan bool)
	return nil
}

func (r *router) createServer() *http.Server {
	muxer := mux.NewRouter()
	muxer.Path("/get_uid").
		Methods(http.MethodGet).
		Schemes("HTTP").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			handleWebRequest(r, w, req)
		})

	address := fmt.Sprintf("%v:%v", r.listenAddress, r.listenPort)
	server := http.Server{
		Addr:    address,
		Handler: muxer,
	}

	go func() {
		log.Infof("Starting server on %v", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Errorf("Error during server startup: %v", err)
		}
	}()
	return &server
}

func handleWebRequest(r *router, w http.ResponseWriter, req *http.Request) {
	// Default timeout, 1 minute
	timeout := 1 * time.Minute

	// All response will be plain text
	w.Header().Set("Content-Type", "text/plain")

	vars := req.URL.Query()
	log.Tracef("handleWebRequest: Received request with parameters: %v", vars)
	if _, exists := vars["timeout"]; exists {
		parsedInt, err := strconv.ParseInt(vars["timeout"][0], 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		timeout = time.Duration(parsedInt) * time.Second
	}

	start := time.Now()
	if err := sendWebRequest(r, timeout); err != nil {
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte(TimeoutError{}.Error()))
		log.Debug("handleWebRequest: Request timed out sending on channel")
		return
	}
	elapsed := time.Since(start)
	log.Tracef("handleWebRequest: sent web request to channel in %v", elapsed)

	// We shouldn't wait the full timeout again, subtrack the time it took to send the web request
	// from the overall timeout
	timeout = time.Duration(timeout.Nanoseconds() - elapsed.Nanoseconds())

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	log.Debugf("handleWebRequest: Waiting for event with timeout: %v", timeout)
	for {
		select {
		case <-timer.C:
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte(TimeoutError{}.Error()))
			log.Debug("handleWebRequest: Request timed out")
			return
		case event := <-r.webChannel:
			log.Debugf("handleWebRequest: Returned '%v'", event.UID())
			w.Write([]byte(event.UID()))
			return
		}
	}
}

func sendWebRequest(r *router, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			return errors.New("Timed out sending web request")
		case r.webRequestChannel <- true:
			return nil
		}
	}
}

func (r *router) handle(event event.Event) error {
	for _, h := range readerctx.SortedHandlers() {
		log.Tracef("%T", h)
		if err := h.Handle(event); err != nil {
			return err
		}
	}
	return nil
}
