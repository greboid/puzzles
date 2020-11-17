package main

import (
	"context"
	"flag"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/csmith/kowalski/v3"
	"github.com/kouhin/envflag"
)

var (
	wordList          = flag.String("wordlist-dir", "/app/wordlists", "Path of the word list directory")
	words             []*kowalski.SpellChecker
)

type Output struct {
	Success bool
	Result  interface{}
}

//go:generate go run flags/flags.go

func main() {
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to parse flags: %s", err.Error())
	}
	log.Printf("Loading wordlist.")
	words = loadWords(*wordList)
	router := httprouter.New()
	router.GET("/anagram", multiplexHandler(kowalski.MultiplexAnagram))
	router.GET("/match", multiplexHandler(kowalski.MultiplexMatch))
	router.GET("/morse", multiplexHandler(kowalski.MultiplexFromMorse))
	router.GET("/t9", multiplexHandler(kowalski.MultiplexFromT9))
	router.GET("/analyse", analyseHandler)
	router.POST("/exifUpload", exifUpload)
	router.NotFound = staticHandler(http.FileServer(http.Dir(filepath.Join(".", "static"))))
	log.Print("Starting server.")
	server := http.Server{
		Addr:    ":8080",
		Handler: requestLogger(router),
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

func staticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.ServeFile(w, r, filepath.Join(".", "static", "404.html"))
			return
		}
		cleanPath := path.Clean(r.URL.Path)
		if _, err :=  os.Stat(filepath.Join(".", "static", cleanPath)); os.IsNotExist(err)  {
			http.ServeFile(w, r, filepath.Join(".", "static", "404.html"))
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

func multiplexHandler(function wordsFunction) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
		input := request.FormValue("input")
		writer.Header().Add("Content-Type", "application/json")
		outputBytes, outputStatus := getResults(words, input, function)
		writer.WriteHeader(outputStatus)
		_, _ = writer.Write(outputBytes)
	}
}

func analyseHandler(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	input := request.FormValue("input")
	writer.Header().Add("Content-Type", "application/json")
	outputBytes, outputStatus := analyse(words[0], input)
	writer.WriteHeader(outputStatus)
	_, _ = writer.Write(outputBytes)
}

func exifUpload(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
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
