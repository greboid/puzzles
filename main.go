package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/csmith/kowalski/v2"
	"github.com/kouhin/envflag"
)

var (
	wordList          = flag.String("wordlist-dir", "/app/wordlists", "Path of the word list directory")
	words             []*kowalski.SpellChecker
	download          = flag.Bool("download-flags", false, "Download new flags data")
)

type Output struct {
	Success bool
	Result  interface{}
}

//go:generate go run . -download-flags

func main() {
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to parse flags: %s", err.Error())
	}
	if *download {
		downloadFlags()
		return
	}
	log.Printf("Loading wordlist.")
	words = loadWords(*wordList)
	mux := http.NewServeMux()
	mux.Handle("/", disableDirectoryListing(http.FileServer(http.Dir(filepath.Join(".", "static")))))
	mux.HandleFunc("/anagram", multiplexHandler(kowalski.MultiplexAnagram))
	mux.HandleFunc("/match", multiplexHandler(kowalski.MultiplexMatch))
	mux.HandleFunc("/morse", multiplexHandler(kowalski.MultiplexFromMorse))
	mux.HandleFunc("/t9", multiplexHandler(kowalski.MultiplexFromT9))
	mux.HandleFunc("/exifUpload", exifUpload)
	log.Print("Starting server.")
	server := http.Server{
		Addr:    ":8080",
		Handler: notFoundHandler(requestLogger(mux)),
	}
	go func() {
		_ = server.ListenAndServe()
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Unable to shutdown: %s", err.Error())
	}
	log.Print("Finishing server.")
}

func serveNotFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(".", "static", "404.html"))
}

func notFoundHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := path.Clean(r.URL.Path)
		if _, err :=  os.Stat(filepath.Join(".", "static", cleanPath)); os.IsNotExist(err)  {
			serveNotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func disableDirectoryListing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			serveNotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetMux.ServeHTTP(w, r)
		requesterIP := r.RemoteAddr
		log.Printf(
			"%s  \t%s  \t%s",
			requesterIP,
			r.Method,
			r.RequestURI,
		)
	})
}

func multiplexHandler(function wordsFunction) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		input := request.FormValue("input")
		writer.Header().Add("Content-Type", "application/json")
		outputBytes, outputStatus := getResults(words, input, function)
		writer.WriteHeader(outputStatus)
		_, _ = writer.Write(outputBytes)
	}
}

func exifUpload(writer http.ResponseWriter, request *http.Request) {
	file, _, err := request.FormFile("exifFile")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("Error"))
		log.Println("Error Getting File", err)
		return
	}
	defer func() {
		_ = file.Close()
	}()
	outputBytes, outputStatus := getImageResults(file)
	writer.WriteHeader(outputStatus)
	_, _ = writer.Write(outputBytes)
}
