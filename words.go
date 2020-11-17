package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/csmith/kowalski/v3"
)

type wordsFunction func([]*kowalski.SpellChecker, string, ...kowalski.MultiplexOption) [][]string

func getResults(checker []*kowalski.SpellChecker, input string, function wordsFunction) (output []byte, statusCode int) {
	if input == "" || len(input) > 100 {
		output, _ = json.Marshal(Output{
			Success: false,
			Result:  "Invalid input",
		})
		statusCode = http.StatusBadRequest
	} else {
		results := function(checker, input, kowalski.Dedupe)
		output, _ = json.Marshal(Output{
			Success: len(results) > 0 && (len(results[0]) > 0 || len(results[1]) > 0),
			Result:  map[string][]string{
				"Standard": results[0],
				"UrbanDictionary": results[1],
			},
		})
		statusCode = http.StatusOK
	}
	return
}

func analyse(words *kowalski.SpellChecker, input string) (output []byte, statusCode int) {
	result := kowalski.Analyse(words, input)
	output, _ = json.Marshal(Output{
		Success: true,
		Result:  result,
	})
	statusCode = http.StatusOK
	return
}

func loadWords(wordlistFolder string) []*kowalski.SpellChecker {
	files, err := ioutil.ReadDir(wordlistFolder)
	if err != nil {
		return nil
	}
	files, err = ioutil.ReadDir(wordlistFolder)
	if err != nil {
		return nil
	}
	var wls []*kowalski.SpellChecker
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".wl") {
			wl, err := loadWordList(filepath.Join(wordlistFolder, file.Name()))
			if err == nil {
				wls = append(wls, wl)
			} else {
				log.Printf("Unable to load wordlist %s: %s", filepath.Join(wordlistFolder, file.Name()+".wl"), err)
			}
		}
	}
	return wls
}

func loadWordList(wordfile string) (*kowalski.SpellChecker, error) {
	log.Printf("Loading wordlist: %s", wordfile)
	f, err := os.Open(wordfile)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()
	return kowalski.LoadSpellChecker(f)
}
