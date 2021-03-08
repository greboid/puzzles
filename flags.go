package main

import (
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"regexp"
	"strings"
)

type flagInfo struct {
	Country     string   `json:"country"`
	Image       string   `json:"image"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

func reduceResult(files fs.FS, input string) (success bool, result []flagInfo) {
	flags, flagKeywords, err := reduceFlags(files)
	if err != nil {
		success = true
		return
	}
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	input = strings.ReplaceAll(input, "-", " ")
	input = regexp.MustCompile(`/[^a-z0-9]/g`).ReplaceAllString(input, "")
	terms := strings.Split(input, " ")
	terms = unique(terms)
	terms = intersects(flagKeywords, terms)
	terms = unique(terms)
	if len(terms) == 0 {
		success = false
		return
	}
	for _, flag := range flags {
		includes := true
		for _, term := range terms {
			if !contains(flag.Keywords, term) {
				includes = false
			}
		}
		if includes {
			result = append(result, flag)
		}
	}
	success = true
	return
}

func unique(input []string) (output []string) {
	if len(input) == 0 {
		return input
	}
	for _, key := range input {
		if !contains(output, key) {
			output = append(output, key)
		}
	}
	return output
}

func contains(input []string, key string) bool {
	for _, k := range input {
		if k == key {
			return true
		}
	}
	return false
}

func intersects(u1 []string, u2 []string) (output []string) {
	if len(u1) > len(u2) {
		for _, t := range u1 {
			if contains(u2, t) {
				output = append(output, t)
			}
		}
	} else {
		for _, t := range u2 {
			if contains(u1, t) {
				output = append(output, t)
			}
		}
	}
	return output
}

func reduceFlags(files fs.FS) ([]flagInfo, []string, error) {
	file, err := files.Open("flags.json")
	if err != nil {
		log.Printf("Unable to open flags.json: %s", err)
		return nil, nil, err
	}
	flagBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Unable to read flags.json: %s", err)
		return nil, nil, err
	}
	flags := make([]flagInfo, 0)
	flagKeywords := make([]string, 0)
	err = json.Unmarshal(flagBytes, &flags)
	if err != nil {
		log.Printf("Unable to unmarshall flags.json: %s", err)
		return nil, nil, err
	}
	for _, flag := range flags {
		for _, flagKeyword := range flag.Keywords {
			flagKeywords = append(flagKeywords, flagKeyword)
		}
	}
	return flags, flagKeywords, nil
}
