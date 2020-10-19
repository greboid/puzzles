package main

import (
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/csmith/kowalski/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/kouhin/envflag"
)

var (
	templates         *template.Template
	wordList          = flag.String("wordlist-dir", "/app/wordlists", "Path of the word list directory")
	templateDirectory = flag.String("template-dir", "/app/templates", "Path of the templates directory")
	words             []*kowalski.SpellChecker
)

type OutputArray struct {
	Success bool
	Result  []string
}

type OutputString struct {
	Success bool
	Result  string
}

func main() {
	err := envflag.Parse()
	if err != nil {
		log.Fatalf("Unable to parse flags: %s", err.Error())
	}
	log.Printf("Loading wordlist.")
	words = loadWords(*wordList)
	log.Print("Loading templates.")
	reloadTemplates()
	templateChanges()
	mux := http.NewServeMux()
	mux.HandleFunc("/static/axios.js", axiosHandler)
	mux.HandleFunc("/favicon.ico", faviconHandler)
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/css", cssHandler)
	mux.HandleFunc("/js", jsHandler)
	mux.HandleFunc("/anagram", anagramHandler)
	mux.HandleFunc("/match", matchHandler)
	mux.HandleFunc("/exifUpload", exifUpload)
	log.Print("Starting server.")
	server := http.Server{
		Addr:    ":8080",
		Handler: requestLogger(mux),
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

func templateChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Print("Unable to create watcher")
		return
	}
	err = watcher.Add(filepath.Join("./templates"))
	if err != nil {
		log.Print("Unable to watch template folder")
	}
	go func() { templateReloader(watcher) }()
}

func templateReloader(watcher *fsnotify.Watcher) {
	for {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				return
			}
			reloadTemplates()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func reloadTemplates() {
	templates = template.Must(template.ParseFiles(
		filepath.Join(*templateDirectory, "main.css"),
		filepath.Join(*templateDirectory, "index.html"),
		filepath.Join(*templateDirectory, "main.js"),
	))
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

func cssHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/css; charset=utf-8")
	err := templates.ExecuteTemplate(writer, "main.css", nil)
	if err != nil {
		log.Printf("Fucked up: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func jsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	err := templates.ExecuteTemplate(writer, "main.js", nil)
	if err != nil {
		log.Printf("Fucked up: %s", err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path == "/" {
		err := templates.ExecuteTemplate(writer, "index.html", "")
		if err != nil {
			log.Printf("Fucked up: %s", err.Error())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func axiosHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(".", "static", "axios.js"))
}

func faviconHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte{})
}

func anagramHandler(writer http.ResponseWriter, request *http.Request) {
	input := request.FormValue("input")
	writer.Header().Add("Content-Type", "application/json")
	outputBytes, outputStatus := getResults(words, input, kowalski.MultiplexAnagram)
	writer.WriteHeader(outputStatus)
	_, _ = writer.Write(outputBytes)
}

func matchHandler(writer http.ResponseWriter, request *http.Request) {
	input := request.FormValue("input")
	writer.Header().Add("Content-Type", "application/json")
	outputBytes, outputStatus := getResults(words, input, kowalski.MultiplexMatch)
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
