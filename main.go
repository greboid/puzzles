package main

import (
	"context"
	"embed"
	"flag"
	"html/template"
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
var staticFiles fs.FS

//go:embed templates
var templateFS embed.FS

var (
	wordList = flag.String("wordlist-dir", "/app/wordlists", "Path of the word list directory")
	words    []*kowalski.SpellChecker
)

//go:generate go run flags/flags.go

func main() {
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to parse flags: %s", err.Error())
	}
	staticFiles, err = GetEmbedOrOSFS("static", staticFS)
	if err != nil {
		log.Fatalf("Unable to get static folder: %s", err.Error())
	}
	templates, err := template.ParseFS(templateFS, "templates/*.tpl")
	if err != nil {
		log.Fatalf("Unable to load templates: %s", err.Error())
	}
	log.Printf("Loading wordlist.")
	words = loadWords(*wordList)
	router := mux.NewRouter()
	router.Use(handlers.ProxyHeaders)
	router.Use(handlers.CompressHandler)
	router.Use(NewLoggingHandler(os.Stdout))
	router.HandleFunc("/anagram", multiplexHandler(kowalski.MultiplexAnagram, templates)).Methods("GET")
	router.HandleFunc("/match", multiplexHandler(kowalski.MultiplexMatch, templates)).Methods("GET")
	router.HandleFunc("/morse", multiplexHandler(kowalski.MultiplexFromMorse, templates)).Methods("GET")
	router.HandleFunc("/t9", multiplexHandler(kowalski.MultiplexFromT9, templates)).Methods("GET")
	router.HandleFunc("/analyse", analyseHandler(templates)).Methods("GET")
	router.HandleFunc("/exifUpload", exifUpload(templates)).Methods("POST")
	router.HandleFunc("/flags", flagResult(templates)).Methods("GET")
	router.PathPrefix("/").Handler(NotFoundHandler(http.FileServer(http.FS(staticFiles)), staticFiles))
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

func multiplexHandler(function wordsFunction, templates *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		input := request.FormValue("input")
		success, results := getResults(words, input, function)
		if !success {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			_ = templates.ExecuteTemplate(writer, "wordlist.tpl", results)
		}
	}
}

func analyseHandler(templates *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		input := request.FormValue("input")
		success, result := analyse(words[0], input)
		if !success {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			_ = templates.ExecuteTemplate(writer, "analysis.tpl", result)
		}
	}
}

func exifUpload(templates *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
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
		success, result := getImageResults(file)
		if !success {
			writer.WriteHeader(http.StatusBadRequest)
		} else {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			_ = templates.ExecuteTemplate(writer, "imageinfo.tpl", result)
		}
	}
}

func flagResult(templates *template.Template) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		input := request.FormValue("input")
		success, result := reduceResult(staticFiles, input)
		if !success {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			_ = templates.ExecuteTemplate(writer, "flags.tpl", nil)
		} else {
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			writer.WriteHeader(http.StatusOK)
			_ = templates.ExecuteTemplate(writer, "flags.tpl", result)
		}
	}
}
