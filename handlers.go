package main

import (
	"github.com/gorilla/handlers"
	"io"
	"net/http"
)

type notFoundInterceptWriter struct {
	realWriter http.ResponseWriter
	status     int
}

func (w *notFoundInterceptWriter) Header() http.Header {
	return w.realWriter.Header()
}

func (w *notFoundInterceptWriter) WriteHeader(status int) {
	w.status = status
	if status != http.StatusNotFound {
		w.realWriter.WriteHeader(status)
	}
}

func (w *notFoundInterceptWriter) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.realWriter.Write(p)
	}
	return len(p), nil
}

func NewLoggingHandler(dst io.Writer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.LoggingHandler(dst, h)
	}
}

func CustomNotFoundHandler(h http.Handler, redirect string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fakeWriter := &notFoundInterceptWriter{realWriter: w}
		h.ServeHTTP(fakeWriter, r)
		if fakeWriter.status == http.StatusNotFound {
			http.Redirect(w, r, redirect, http.StatusFound)
		}
	}
}

func NotFoundHandler(h http.Handler) http.HandlerFunc {
	return CustomNotFoundHandler(h, "/404.html")
}
