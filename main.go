package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/csmith/kowalski"
	"github.com/fsnotify/fsnotify"
	"github.com/kouhin/envflag"
	"github.com/simpicapp/goexif/exif"
	"github.com/simpicapp/goexif/tiff"
)

var (
	templates         *template.Template
	wordList          = flag.String("word-list", "/app/wordlist.txt", "Path of the word list file")
	templateDirectory = flag.String("template-dir", "/app/templates", "Path of the templates directory")
	words             *kowalski.Node
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
	words, err = loadWords(*wordList)
	if err != nil {
		log.Printf("Unable to load words: %s", err.Error())
	}
	log.Print("Loading templates.")
	reloadTemplates()
	templateChanges()
	mux := http.NewServeMux()
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
			"%s\t\t%s\t\t%s\t",
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

func anagramHandler(writer http.ResponseWriter, request *http.Request) {
	input := request.FormValue("input")
	writer.Header().Add("Content-Type", "application/json")
	outputBytes, outputStatus := getResults(input, words.Anagrams)
	writer.WriteHeader(outputStatus)
	_, _ = writer.Write(outputBytes)
}

func matchHandler(writer http.ResponseWriter, request *http.Request) {
	input := request.FormValue("input")
	writer.Header().Add("Content-Type", "application/json")
	outputBytes, outputStatus := getResults(input, words.Match)
	writer.WriteHeader(outputStatus)
	_, _ = writer.Write(outputBytes)
}

func getResults(input string, function func(string) []string) (output []byte, statusCode int) {
	if input == "" || len(input) > 13 {
		output, _ = json.Marshal(OutputString{
			Success: false,
			Result:  "Invalid input",
		})
		statusCode = http.StatusBadRequest
	} else {
		result := function(input)
		output, _ = json.Marshal(&OutputArray{
			Success: len(result) > 0,
			Result:  result,
		})
		statusCode = http.StatusOK
	}
	return
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
	exifData, err := getExifData(file)
	if err != nil {
		if err == exif.NotFoundError || err == io.EOF {
			output, _ := json.Marshal(OutputString{
				Success: false,
				Result:  "[]",
			})
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write(output)
			return
		}
		output, _ := json.Marshal(OutputString{
			Success: false,
			Result:  "Error parsing EXIF",
		})
		log.Println("Error Parsing EXIF", err)
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write(output)
		return
	}
	output, _ := json.Marshal(parseExif(exifData))
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(output)
}

func getExifData(input io.Reader) (*exif.Exif, error) {
	exifData, err := exif.Decode(input)
	if err != nil {
		return nil, err
	}
	return exifData, nil
}

func parseExif(exifData *exif.Exif) *OutputString {
	var data []string

	values := make(map[string]string)
	walker := &walker{values}
	_ = exifData.Walk(walker)
	for key, value := range values {
		data = append(data, key+": "+value)
	}
	sort.Strings(data)
	data = append([]string{"----Raw Values----"}, data...)
	datetime, err := exifData.DateTime()
	if err == nil {
		data = append([]string{"Date: " + datetime.String()}, data...)
	}
	lat, long, err := exifData.LatLong()
	if err == nil {
		data = append([]string{fmt.Sprintf("Maps Link: https://www.google.com/maps/search/?api=1&query=%f,%f", lat, long)}, data...)
	}
	comment, err := exifData.Get("usercomment")
	if err == nil {
		data = append([]string{"Comment: " + comment.String()}, data...)
	}
	result, _ := json.Marshal(data)
	return &OutputString{
		Success: true,
		Result:  string(result),
	}
}

type walker struct {
	values map[string]string
}

func (e *walker) Walk(name exif.FieldName, tag *tiff.Tag) error {
	e.values[string(name)] = tag.String()
	return nil
}

func Walk(name exif.FieldName, tag *tiff.Tag) error {
	return nil
}

func loadWords(wordfile string) (*kowalski.Node, error) {
	if _, err := os.Stat(wordfile + ".gob"); err == nil {
		log.Printf("Using cached wordlist")
		wordfile = wordfile + ".gob"
	}
	if strings.HasSuffix(wordfile, ".gob") {
		f, err := os.Open(wordfile)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = f.Close()
		}()
		words = &kowalski.Node{}
		if err := gob.NewDecoder(f).Decode(&words); err != nil {
			return nil, err
		}
		return words, nil
	}
	words, err := kowalski.LoadWords(wordfile)
	if err != nil {
		return nil, err
	}
	return words, nil
}
