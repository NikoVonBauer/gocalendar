package main

import (
	"net/http"
	"encoding/json"
	"os"
	"path/filepath"
	"html/template"

	"log"
)

// Структура записей
type Entry struct {
	Date string `json:"date"`
	Content string `json:"content"`
}

// Загрузка записей из data/diary.json
func loadEntries() ([]Entry, error) {
	filePath := filepath.Join("data", "diary.json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []Entry
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&entries)
	return entries, err
}
// Метод сохранения записей (формализация)
func saveEntries(entries []Entry) error {
	filePath := filepath.Join("data", "diary.json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}
// Base endpoint
func indexHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := loadEntries()
	if err != nil {
		http.Error(w, "Не удалось загрузить записи", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, entries)
}
// Add new entry
func addHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	date := r.FormValue("date")
	content := r.FormValue("content")

	if date == "" || content == "" {
		http.Error(w, "Заполните все поля", http.StatusBadRequest)
		return
	}

	newEntry := Entry{Date: date, Content: content}

	entries, err := loadEntries()
	if err != nil {
		http.Error(w, "Ошибка при загрузке данных", http.StatusInternalServerError)
		return
	}

	entries = append([]Entry{newEntry}, entries...)

	if err := saveEntries(entries); err != nil {
		http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?success=added", http.StatusSeeOther)
}
// Delete existing entry
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	date := r.FormValue("date")
	content := r.FormValue("content")

	entries, err := loadEntries()
	if err != nil {
		http.Error(w, "Ошибка при загрузке данных", http.StatusInternalServerError)
		return
	}

	newEntries := make([]Entry, 0)
	for _, e := range entries {
		if e.Date != date || e.Content != content {
			newEntries = append(newEntries, e)
		}
	}

	if err := saveEntries(newEntries); err != nil {
		http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?success=deleted", http.StatusSeeOther)
}
// Edit existing handler (GET request for pick existing entry)
func editGetHandler(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	content := r.URL.Query().Get("content")

	entry := Entry{
		Date:    date,
		Content: content,
	}

	tmpl, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, entry)
}
// Edit existing handler (POST request for saving new entry)
func editPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	originalDate := r.FormValue("originalDate")
	originalContent := r.FormValue("originalContent")
	newDate := r.FormValue("date")
	newContent := r.FormValue("content")
	entries, err := loadEntries()
	if err != nil {
		http.Error(w, "Ошибка при загрузке данных", http.StatusInternalServerError)
		return
	}

	for i, e := range entries {
		if e.Date == originalDate && e.Content == originalContent {
			entries[i] = Entry{Date: newDate, Content: newContent}
			break
		}
	}

	if err := saveEntries(entries); err != nil {
		http.Error(w, "Ошибка при сохранении данных", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/?success=edited", http.StatusSeeOther)
}


func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/add", addHandler)
	mux.HandleFunc("/delete", deleteHandler)
	mux.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			editGetHandler(w, r)
		case http.MethodPost:
			editPostHandler(w, r)
		default:
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

