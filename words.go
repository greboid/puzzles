package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/csmith/kowalski/v2"
)

func getResults(checker *kowalski.SpellChecker, input string, function func(*kowalski.SpellChecker, string) []string) (output []byte, statusCode int) {
	if input == "" || len(input) > 13 {
		output, _ = json.Marshal(OutputString{
			Success: false,
			Result:  "Invalid input",
		})
		statusCode = http.StatusBadRequest
	} else {
		result := function(checker, input)
		output, _ = json.Marshal(&OutputArray{
			Success: len(result) > 0,
			Result:  result,
		})
		statusCode = http.StatusOK
	}
	return
}

func loadWords(wordfile string) (*kowalski.SpellChecker, error) {
	if _, err := os.Stat(wordfile + ".wl"); err == nil {
		log.Printf("Cached spellchecker found, using")
		wordfile = wordfile + ".wl"
	}
	if !strings.HasSuffix(wordfile, ".wl") {
		log.Printf("Creating cached spellchecker")
		if _, err := os.Stat(wordfile + ".wl"); err != nil {
			f, err := os.Open(wordfile)
			if err != nil {
				return nil, err
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				return nil, err
			}
			count := bytes.Count(b, []byte{'\n'})
			words, err := kowalski.CreateSpellChecker(bytes.NewReader(b), count)
			if err != nil {
				return nil, err
			}
			_ = f.Close()
			w, err := os.Create(wordfile + ".wl")
			err = kowalski.SaveSpellChecker(w, words)
			if err != nil {
				return nil, err
			}
			_ = w.Close()
		}
		wordfile = wordfile + ".wl"
	}
	f, err := os.Open(wordfile)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return kowalski.LoadSpellChecker(f)
}
