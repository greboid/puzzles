package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

const imagesFolder = "images/webp-resized/"
const flagsJson = "flags.json"

func main() {
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: "https://github.com/csmith/flagdata",
	})
	log.Printf("Getting git repo")
	CheckIfError("cloning flags data", err)
	head, err := r.Head()
	CheckIfError("retrieving head ", err)
	commit, err := r.CommitObject(head.Hash())
	CheckIfError("getting commit", err)
	tree, err := commit.Tree()
	CheckIfError("getting git tree", err)
	log.Printf("Downloading flags.json")
	flagsData, err := tree.File(flagsJson)
	CheckIfError("getting flags.json file", err)
	flagsReader, err := flagsData.Reader()
	CheckIfError("getting flags.json file reader", err)
	body, err := ioutil.ReadAll(flagsReader)
	CheckIfError("getting reading flags.json", err)
	_ = flagsReader.Close()
	err = ioutil.WriteFile(filepath.Join(".", "static", "static", "flags.json"), body, 0777)
	CheckIfError("writing flags.json", err)
	log.Printf("Downloaded flags.json")
	fileIter := tree.Files()
	log.Printf("Downloading flag images")
	err = fileIter.ForEach(func(file *object.File) error {
		if strings.HasPrefix(file.Name, imagesFolder) {
			filename := strings.TrimPrefix(file.Name, imagesFolder)
			reader, err := file.Reader()
			CheckIfError(fmt.Sprintf("error getting reader for file %s", filename), err)
			body,err := ioutil.ReadAll(reader)
			CheckIfError(fmt.Sprintf("error reading file %s", filename), err)
			err = ioutil.WriteFile(filepath.Join(".", "static", "static", "flags", filename), body, 0777)
			CheckIfError(fmt.Sprintf("error writing file %s", filename), err)
			_ = reader.Close()
			log.Printf("Downloaded %s", filename)
		}
		return nil
	})
	CheckIfError("getting flags", err)
	log.Printf("Flags data downloaded.")
}

func CheckIfError(desc string, err error) {
	if err != nil {
		log.Fatalf("Error whilst %s: %s", desc, err.Error())
	}
}