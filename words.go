package main

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/csmith/kowalski"
)

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
