package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/ncruces/go-sqlite3"
)

var db *sql.DB
var homeTemplate *template.Template
var entriesTemplate *template.Template

type Entry struct {
	ID    int
	Title string
	Text  string
}

func init() {
	var err error
	db, err = sql.Open("sqlite3", "entries.db")
	if err != nil {
		log.Fatal(err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		text TEXT NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}

	var parseErr error
	homeTemplate, parseErr = template.ParseFiles("templates/home.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}

	entriesTemplate, parseErr = template.ParseFiles("templates/entries.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}
}

func main() {
	defer db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/entries", entriesHandler)
	http.HandleFunc("/add-entry", addEntryHandler)

	addr := ":8080"
	fmt.Println("Starting server on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := homeTemplate.Execute(w, nil); err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func addEntryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	text := strings.TrimSpace(r.FormValue("text"))

	if title == "" || text == "" {
		http.Error(w, "Title and text are required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO entries (title, text) VALUES (?, ?)", title, text)
	if err != nil {
		http.Error(w, "Failed to add entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderEntries(w)
}

func entriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	renderEntries(w)
}

func renderEntries(w http.ResponseWriter) {
	rows, err := db.Query("SELECT id, title, text FROM entries ORDER BY id DESC")
	if err != nil {
		fmt.Fprint(w, `<p style="color: red;">Error loading entries</p>`)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.Title, &entry.Text); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	if err := entriesTemplate.Execute(w, entries); err != nil {
		log.Println("Error executing entries template:", err)
		fmt.Fprint(w, `<p style="color: red;">Error rendering entries</p>`)
	}
}