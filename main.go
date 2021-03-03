package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/csmith/kowalski/v3"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kouhin/envflag"
)

//go:embed static
var staticFS embed.FS

var (
	wordList = flag.String("wordlist-dir", "/app/wordlists", "Path of the word list directory")
	words    []*kowalski.SpellChecker
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
	staticFiles, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Unable to get static folder: %s", err.Error())
	}
	log.Printf("Loading wordlist.")
	words = loadWords(*wordList)
	router := mux.NewRouter()
	router.Use(handlers.ProxyHeaders)
	router.Use(handlers.CompressHandler)
	router.Use(NewLoggingHandler(os.Stdout))
	router.HandleFunc("/anagram", multiplexHandler(kowalski.MultiplexAnagram)).Methods("GET")
	router.HandleFunc("/match", multiplexHandler(kowalski.MultiplexMatch)).Methods("GET")
	router.HandleFunc("/morse", multiplexHandler(kowalski.MultiplexFromMorse)).Methods("GET")
	router.HandleFunc("/t9", multiplexHandler(kowalski.MultiplexFromT9)).Methods("GET")
	router.HandleFunc("/analyse", analyseHandler).Methods("GET")
	router.HandleFunc("/exifUpload", exifUpload).Methods("POST")
	router.PathPrefix("/").Handler(NotFoundHandler(http.FileServer(http.FS(staticFiles))))
	log.Print("Starting server.")
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
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

func multiplexHandler(function wordsFunction) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		input := request.FormValue("input")
		writer.Header().Add("Content-Type", "application/json")
		outputBytes, outputStatus := getResults(words, input, function)
		writer.WriteHeader(outputStatus)
		_, _ = writer.Write(outputBytes)
	}
}

func analyseHandler(writer http.ResponseWriter, request *http.Request) {
	input := request.FormValue("input")
	writer.Header().Add("Content-Type", "application/json")
	outputBytes, outputStatus := analyse(words[0], input)
	writer.WriteHeader(outputStatus)
	_, _ = writer.Write(outputBytes)
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
