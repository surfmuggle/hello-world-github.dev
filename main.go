package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var homeTemplate *template.Template
var entriesTemplate *template.Template
var questionnaireTemplate *template.Template

type Entry struct {
	ID          int
	Title       string
	Text        string
	Toolname    string
	CreatedDate string
	UpdatedDate string
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
		text TEXT NOT NULL,
		toolname TEXT NOT NULL,
		created_date DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_date DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal(err)
	}

	questionsSchema := `
	CREATE TABLE IF NOT EXISTS questions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data TEXT NOT NULL,
		created_date DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(questionsSchema); err != nil {
		log.Fatal(err)
	}

	var parseErr error
	homeTemplate, parseErr = template.ParseFiles("templates/home.html", "templates/nav.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}

	entriesTemplate, parseErr = template.ParseFiles("templates/entries.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}

	questionnaireTemplate, parseErr = template.ParseFiles("templates/questionnaire.html", "templates/nav.html")
	if parseErr != nil {
		log.Fatal(parseErr)
	}
}

func main() {
	defer db.Close()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/entries", entriesHandler)
	http.HandleFunc("/add-entry", addEntryHandler)
	http.HandleFunc("/questionnaire", questionnaireHandler)
	http.HandleFunc("/submit-questionnaire", submitQuestionnaireHandler)

	addr := ":8080"
	fmt.Println("Starting server on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

type RawField struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	FormType    string   `json:"formtype"`
	Name        string   `json:"name"`
	Options     []string `json:"options,omitempty"`
}

func questionnaireHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// read JSON definition
	b, err := os.ReadFile("data/questions.json")
	if err != nil {
		http.Error(w, "Failed to load form definition", http.StatusInternalServerError)
		return
	}

	var fields []RawField
	if err := json.Unmarshal(b, &fields); err != nil {
		http.Error(w, "Invalid form definition", http.StatusInternalServerError)
		return
	}

	// build safe HTML snippets
	var parts []template.HTML
	for _, f := range fields {
		name := f.Name
		if name == "" {
			// fallback name
			name = strings.ReplaceAll(strings.ToLower(f.Title), " ", "_")
		}
		var html string
		switch strings.ToLower(f.FormType) {
		case "textarea", "texarea":
			html = fmt.Sprintf(`<div class="form-group"><label>%s</label><textarea name="%s"></textarea><p class="desc">%s</p></div>`, template.HTMLEscapeString(f.Title), template.HTMLEscapeString(name), template.HTMLEscapeString(f.Description))
		case "select":
			opts := ""
			for _, o := range f.Options {
				opts += fmt.Sprintf(`<option value="%s">%s</option>`, template.HTMLEscapeString(o), template.HTMLEscapeString(o))
			}
			html = fmt.Sprintf(`<div class="form-group"><label>%s</label><select name="%s">%s</select><p class="desc">%s</p></div>`, template.HTMLEscapeString(f.Title), template.HTMLEscapeString(name), opts, template.HTMLEscapeString(f.Description))
		case "input radio", "radio":
			opts := ""
			for i, o := range f.Options {
				opts += fmt.Sprintf(`<label><input type="radio" name="%s" value="%s"> %s</label> `, template.HTMLEscapeString(name), template.HTMLEscapeString(o), template.HTMLEscapeString(o))
				_ = i
			}
			html = fmt.Sprintf(`<div class="form-group"><div class="label">%s</div>%s<p class="desc">%s</p></div>`, template.HTMLEscapeString(f.Title), opts, template.HTMLEscapeString(f.Description))
		case "input checkbox", "checkbox":
			opts := ""
			for _, o := range f.Options {
				opts += fmt.Sprintf(`<label><input type="checkbox" name="%s" value="%s"> %s</label> `, template.HTMLEscapeString(name), template.HTMLEscapeString(o), template.HTMLEscapeString(o))
			}
			html = fmt.Sprintf(`<div class="form-group"><div class="label">%s</div>%s<p class="desc">%s</p></div>`, template.HTMLEscapeString(f.Title), opts, template.HTMLEscapeString(f.Description))
		default:
			// input text
			html = fmt.Sprintf(`<div class="form-group"><label>%s</label><input type="text" name="%s"><p class="desc">%s</p></div>`, template.HTMLEscapeString(f.Title), template.HTMLEscapeString(name), template.HTMLEscapeString(f.Description))
		}
		parts = append(parts, template.HTML(html))
	}

	data := struct {
		Fields []template.HTML
	}{Fields: parts}

	if err := questionnaireTemplate.Execute(w, data); err != nil {
		log.Println("Error executing questionnaire template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func submitQuestionnaireHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	result := make(map[string]interface{})
	for k, vals := range r.PostForm {
		if len(vals) == 1 {
			result[k] = vals[0]
		} else {
			result[k] = vals
		}
	}

	b, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Failed to encode submission", http.StatusInternalServerError)
		return
	}

	if _, err := db.Exec("INSERT INTO questions (data, created_date) VALUES (?, CURRENT_TIMESTAMP)", string(b)); err != nil {
		http.Error(w, "Failed to store submission", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<p>Danke — Ihre Antwort wurde gespeichert.</p><p><a href=\"/questionnaire\">Zurück</a></p>")
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
	toolname := strings.TrimSpace(r.FormValue("toolname"))

	if title == "" || text == "" || toolname == "" {
		http.Error(w, "Title, text, and toolname are required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("INSERT INTO entries (title, text, toolname, created_date, updated_date) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", title, text, toolname)
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
	rows, err := db.Query("SELECT id, title, text, toolname, created_date, updated_date FROM entries ORDER BY id DESC")
	if err != nil {
		fmt.Fprint(w, `<p style="color: red;">Error loading entries</p>`)
		return
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.Title, &entry.Text, &entry.Toolname, &entry.CreatedDate, &entry.UpdatedDate); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	if err := entriesTemplate.Execute(w, entries); err != nil {
		log.Println("Error executing entries template:", err)
		fmt.Fprint(w, `<p style="color: red;">Error rendering entries</p>`)
	}
}
