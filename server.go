package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

type Node struct {
	ID         int64   `json:"id"`
	Title      string  `json:"title"`
	ChildNodes []*Node `json:"childNodes"`
}

const dataFile = "./JSON/treeData.json"

var treeMutex = new(sync.Mutex)

// Handle Tree
func handleTree(w http.ResponseWriter, r *http.Request) {
	// Since multiple requests could come in at once, ensure we have a lock
	// around all file operations
	treeMutex.Lock()
	defer treeMutex.Unlock()

	// Stat the file, so we can find its current permissions
	fi, err := os.Stat(dataFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to stat the data file (%s): %s", dataFile, err), http.StatusInternalServerError)
		return
	}
	// Read the tree from the file.
	treeData, err := ioutil.ReadFile(dataFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to read the data file (%s): %s", dataFile, err), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case "POST":
		// Decode the JSON data
		var Tree []Node
		if err := json.Unmarshal(treeData, &Tree); err != nil {
			http.Error(w, fmt.Sprintf("Unable to Unmarshal Tree from data file (%s): %s", dataFile, err), http.StatusInternalServerError)
			return
		}

		// Processing data
		// code //

		// Marshal the tree to indented json.
		treeData, err = json.MarshalIndent(Tree, "", "    ")
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to marshal Tree to json: %s", err), http.StatusInternalServerError)
			return
		}

		// Write out the tree to the file, preserving permissions
		err := ioutil.WriteFile(dataFile, treeData, fi.Mode())
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to write tree to data file (%s): %s", dataFile, err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		io.Copy(w, bytes.NewReader(treeData))

	case "GET":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// stream the contents of the file to the response
		io.Copy(w, bytes.NewReader(treeData))

	default:
		// Don't know the method, so error
		http.Error(w, fmt.Sprintf("Unsupported method: %s", r.Method), http.StatusMethodNotAllowed)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	http.HandleFunc("/api/treeData", handleTree)
	// http.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Server started: http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
