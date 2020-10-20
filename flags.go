package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

func downloadFlags() {
	flagsURL := "https://ghcdn.rawgit.org/csmith/flagdata/master/flags.json"
	rsp, err := http.Get(flagsURL)
	if err != nil {
		log.Printf("Unable to download flags: %s", err.Error())
		return
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Printf("Unable to read flags: %s", err.Error())
		return
	}
	_ = rsp.Body.Close()
	err = ioutil.WriteFile(filepath.Join(".", "static", "flags.json"), body, 0777)
	if err != nil {
		log.Printf("Unable to download flags: %s", err.Error())
		return
	}
	log.Printf("Flags data downloaded.")
}