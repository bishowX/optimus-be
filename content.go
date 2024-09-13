package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

var contents []Content = []Content{}

func saveContents(content Content) {
	contents = append(contents, content)
	contentsFile, _ := json.Marshal(contents)
	_ = os.WriteFile("contents.json", contentsFile, 0644)
}

func loadContents() {
	contentsFile, err := os.Open("contents.json")
	if err == nil {
		defer contentsFile.Close()
		byteValue, _ := io.ReadAll(contentsFile)
		json.Unmarshal(byteValue, &contents)
	}
}

type Content struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Snippet     string `json:"snippet"`
	Content     string `json:"content"`
	Author      string `json:"author"`
	PublishedOn string `json:"published_on"`
	UpdatedOn   string `json:"updated_on"`
}

func createContent(w http.ResponseWriter, r *http.Request) {
	// read and parse the request body into a Content struct
	var content Content
	err := json.NewDecoder(r.Body).Decode(&content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate the content
	if content.Title == "" || content.Content == "" || content.Author == "" {
		http.Error(w, "title, content, and author are required fields", http.StatusBadRequest)
		return
	}

	// create the content
	content.Id = uuid.New().String()
	content.PublishedOn = time.Now().Format(time.RFC3339)
	content.UpdatedOn = time.Now().Format(time.RFC3339)

	saveContents(content)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(content)
}

func getContentById(id string) (Content, error) {
	for _, content := range contents {
		if content.Id == id {
			return content, nil
		}
	}
	return Content{}, errors.New("content not found")
}
func getContent(w http.ResponseWriter, r *http.Request) {
	// get the content id from the request path parameter not query string
	contentId := r.PathValue("id")

	// get the content from the file system
	content, err := getContentById(contentId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// return the content
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(content)
}
