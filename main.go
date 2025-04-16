package main

import (
	"bufio"
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

type Message struct {
	Time    string `json:"time"`
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

var (
	templates = template.Must(template.ParseFiles("templates/index.html"))
	messages  []Message
	lock      sync.Mutex
	yourName  = "~KaKarot" // or your actual name in chat
)

func main() {
	http.HandleFunc("/", uploadForm)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/messages", loadMessages)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.ListenAndServe(":8080", nil)
}

func uploadForm(w http.ResponseWriter, r *http.Request) {
	templates.Execute(w, nil)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("chatfile")
	if err != nil {
		http.Error(w, "File error", http.StatusBadRequest)
		return
	}
	defer file.Close()

	lock.Lock()
	defer lock.Unlock()
	messages = nil

	scanner := bufio.NewScanner(file)
	regex := regexp.MustCompile(`^\[(.*?)\] (.*?): (.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		match := regex.FindStringSubmatch(line)
		if len(match) == 4 {
			sender := match[2]
			if sender == yourName {
				sender = "You"
			}
			messages = append(messages, Message{
				Time:    match[1],
				Sender:  sender,
				Content: match[3],
			})
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loadMessages(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	pageSize := 20
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	start := page * pageSize
	end := start + pageSize

	if start > len(messages) {
		start = len(messages)
	}
	if end > len(messages) {
		end = len(messages)
	}

	json.NewEncoder(w).Encode(messages[start:end])
}
