package main

import (
	"bufio"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type Message struct {
	Time    string
	Sender  string
	Content string
}

var templates = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", uploadForm)
	http.HandleFunc("/upload", handleUpload)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func uploadForm(w http.ResponseWriter, r *http.Request) {
	templates.Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB limit
	file, handler, err := r.FormFile("chatfile")
	if err != nil {
		http.Error(w, "File error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempPath := filepath.Join(os.TempDir(), handler.Filename)
	tempFile, _ := os.Create(tempPath)
	defer tempFile.Close()
	scanner := bufio.NewScanner(file)

	var messages []Message
	regex := regexp.MustCompile(`^\[(.*?)\] (.*?): (.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		match := regex.FindStringSubmatch(line)
		if len(match) == 4 {
			messages = append(messages, Message{
				Time:    match[1],
				Sender:  match[2],
				Content: match[3],
			})
		}
	}

	templates.Execute(w, messages)
}
