package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

// Page ...
type Page struct {
	Title string
	Body  []byte
	HTML  template.HTML
}

var (
	templates = template.Must(template.ParseGlob("templates/*.html"))
	validPage = regexp.MustCompile("([A-Z][a-z]+[A-Z][a-zA-Z]+)")
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

func (p *Page) save() error {
	filename := path.Join("data", p.Title+".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := path.Join("data", title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Process and Parse the Markdown content
	// Also automatically replace CamelCase page identifiers as links
	markdown := validPage.ReplaceAll(
		body,
		[]byte("[$1](http://localhost/view/$1)"),
	)

	unsafe := blackfriday.MarkdownCommon(markdown)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	return &Page{
		Title: title,
		Body:  body,
		HTML:  template.HTML(html),
	}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	if strings.TrimPrefix(title, "/") == "" {
		http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
		return
	}
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// WikiHandler ...
type WikiHandler func(http.ResponseWriter, *http.Request, string)

func makeHandler(fn WikiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.Handle("/", http.RedirectHandler("/view/FrontPage", http.StatusFound))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))

	css := http.StripPrefix("/css/", http.FileServer(http.Dir("static/css/")))
	http.Handle("/css/", css)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
