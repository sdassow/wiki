package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"

	// Logging
	"github.com/unrolled/logger"

	// Stats/Metrics
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/thoas/stats"

	"github.com/GeertJohan/go.rice"
	"github.com/julienschmidt/httprouter"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var (
	validPage = regexp.MustCompile("([A-Z][a-z]+[A-Z][a-zA-Z]+)")
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

// Page ...
type Page struct {
	Title string
	Body  []byte
	HTML  template.HTML
}

func (p *Page) Save(datadir string) error {
	filename := path.Join(datadir, p.Title+".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// LoadPage ...
func LoadPage(title string, datadir string, baseurl *url.URL) (*Page, error) {
	filename := path.Join(datadir, title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Process and Parse the Markdown content
	// Also automatically replace CamelCase page identifiers as links
	markdown := validPage.ReplaceAll(
		body,
		[]byte(fmt.Sprintf("[$1](%s$1)", baseurl.String())),
	)

	unsafe := blackfriday.MarkdownCommon(markdown)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	return &Page{
		Title: title,
		Body:  body,
		HTML:  template.HTML(html),
	}, nil
}

// Counters ...
type Counters struct {
	r metrics.Registry
}

func NewCounters() *Counters {
	counters := &Counters{
		r: metrics.NewRegistry(),
	}
	return counters
}

func (c *Counters) Inc(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(1)
}

func (c *Counters) Dec(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(1)
}

func (c *Counters) IncBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(n)
}

func (c *Counters) DecBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(n)
}

// Server ...
type Server struct {
	bind      string
	config    Config
	templates *Templates
	router    *httprouter.Router

	// Logger
	logger *logger.Logger

	// Stats/Metrics
	counters *Counters
	stats    *stats.Stats
}

func (s *Server) render(name string, w http.ResponseWriter, ctx interface{}) {
	buf, err := s.templates.Exec(name, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// IndexHandler ...
func (s *Server) IndexHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_index")

		u, err := url.Parse(fmt.Sprintf("./view/FrontPage"))
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}

		http.Redirect(w, r, r.URL.ResolveReference(u).String(), http.StatusFound)
	}
}

// EditHandler ...
func (s *Server) EditHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_edit")

		title := p.ByName("title")

		u, err := url.Parse("../view/")
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		baseurl := r.URL.ResolveReference(u)

		page, err := LoadPage(title, s.config.data, baseurl)
		if err != nil {
			page = &Page{Title: title}
		}

		s.render("edit", w, page)
	}
}

// SaveHandler ...
func (s *Server) SaveHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_save")

		title := p.ByName("title")

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		body := r.Form.Get("body")

		page := &Page{Title: title, Body: []byte(body)}
		err = page.Save(s.config.data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u, err := url.Parse(fmt.Sprintf("/view/%s", title))
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}

		http.Redirect(w, r, r.URL.ResolveReference(u).String(), http.StatusFound)
	}
}

// ViewHandler ...
func (s *Server) ViewHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_view")

		title := p.ByName("title")

		u, err := url.Parse("../view/")
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		baseurl := r.URL.ResolveReference(u)

		page, err := LoadPage(title, s.config.data, baseurl)
		if err != nil {
			u, err := url.Parse(fmt.Sprintf("../edit/%s", title))
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
			}

			http.Redirect(
				w, r, r.URL.ResolveReference(u).String(), http.StatusFound,
			)

			return
		}

		s.render("view", w, page)
	}
}

// StatsHandler ...
func (s *Server) StatsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bs, err := json.Marshal(s.stats.Data())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
	}
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(
		http.ListenAndServe(
			s.bind,
			s.logger.Handler(
				s.stats.Handler(s.router),
			),
		),
	)
}

func (s *Server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.StatsHandler())

	s.router.ServeFiles(
		"/css/*filepath",
		rice.MustFindBox("static/css").HTTPBox(),
	)

	s.router.ServeFiles(
		"/js/*filepath",
		rice.MustFindBox("static/js").HTTPBox(),
	)

	s.router.GET("/", s.IndexHandler())
	s.router.GET("/view/:title", s.ViewHandler())
	s.router.GET("/edit/:title", s.EditHandler())
	s.router.POST("/save/:title", s.SaveHandler())
}

// NewServer ...
func NewServer(bind string, config Config) *Server {
	server := &Server{
		bind:      bind,
		config:    config,
		router:    httprouter.New(),
		templates: NewTemplates("base"),

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "wiki",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
			OutputFlags:          log.LstdFlags,
		}),

		// Stats/Metrics
		counters: NewCounters(),
		stats:    stats.New(),
	}

	// Templates
	box := rice.MustFindBox("templates")

	editTemplate := template.New("view")
	template.Must(editTemplate.Parse(box.MustString("edit.html")))
	template.Must(editTemplate.Parse(box.MustString("base.html")))

	viewTemplate := template.New("view")
	template.Must(viewTemplate.Parse(box.MustString("view.html")))
	template.Must(viewTemplate.Parse(box.MustString("base.html")))

	server.templates.Add("edit", editTemplate)
	server.templates.Add("view", viewTemplate)

	/*
		err := server.templates.Load()
		if err != nil {
			log.Panicf("error loading templates: %s", err)
		}
	*/

	server.initRoutes()

	return server
}
