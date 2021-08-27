package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/config"
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
}

func NewRouter() (*router, error) {
	router := router{}
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
		return nil
	default:
		// There is no web request waiting, send to handlers
		r.handle(event)
		log.Tracef("Default Route complete, state: %#v", readerctx.State)
		return nil
	}
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

	address := fmt.Sprintf("localhost:%v", config.PortNumber)
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

	// Indicate there is a web requst waiting
	r.webRequestChannel <- true

	// All response will be plain text
	w.Header().Set("Content-Type", "text/plain")

	vars := req.URL.Query()
	log.Debugf("Received request with parameters: %v", vars)
	if _, exists := vars["timeout"]; exists {
		parsedInt, err := strconv.ParseInt(vars["timeout"][0], 0, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		timeout = time.Duration(parsedInt) * time.Second
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	log.Debugf("Getting UID with timeout: %v", timeout)
	for {
		select {
		case <-timer.C:
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte(TimeoutError{}.Error()))
			return
		case event := <-r.webChannel:
			w.Write([]byte(event.UID()))
			return
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
