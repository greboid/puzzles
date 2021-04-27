package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/csmith/kowalski/v5"
)

type wordsFunctionWithContext func(context.Context, []*kowalski.SpellChecker, string, ...kowalski.MultiplexOption) ([][]string, error)
type wordsFunction func([]*kowalski.SpellChecker, string, ...kowalski.MultiplexOption) [][]string

func getResultsWithContext(checker []*kowalski.SpellChecker, input string, function wordsFunctionWithContext) (success bool, results map[string][]string) {
	if input == "" || len(input) > 100 {
		success = false
		results = make(map[string][]string)
	} else {
		success = true
		tmpResults, err := function(context.Background(), checker, input, kowalski.Dedupe)
		if err != nil {
			success = false
		}
		results = map[string][]string{"Standard": tmpResults[0], "Urban Dictionary": tmpResults[1]}
	}
	return
}


func getResults(checker []*kowalski.SpellChecker, input string, function wordsFunction) (success bool, results map[string][]string) {
	if input == "" || len(input) > 100 {
		success = false
		results = make(map[string][]string)
	} else {
		success = true
		tmpResults := function(checker, input, kowalski.Dedupe)
		results = map[string][]string{"Standard": tmpResults[0], "Urban Dictionary": tmpResults[1]}
	}
	return
}

func analyse(words *kowalski.SpellChecker, input string) (success bool, result []string) {
	success = true
	result = kowalski.Analyse(words, input)
	return
}

func loadWords(wordlistFolder string) []*kowalski.SpellChecker {
	files, err := os.ReadDir(wordlistFolder)
	if err != nil {
		return nil
	}
	files, err = os.ReadDir(wordlistFolder)
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
