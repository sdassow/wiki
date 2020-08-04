package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	// Logging
	"github.com/unrolled/logger"

	// Stats/Metrics
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/thoas/stats"

	"github.com/GeertJohan/go.rice"
	"github.com/julienschmidt/httprouter"
	"github.com/microcosm-cc/bluemonday"
)

var (
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

// Page ...
type Page struct {
	Title string
	Body  []byte
	HTML  template.HTML
	Brand string
	Date  time.Time
	Files []ListFile
}

// make sure user input path does not leave the directory
func mkSubDir(dir string, file string) error {
	d := path.Clean(dir)
	sd := path.Dir(path.Clean(path.Join(d, file)))
	if sd[0:len(d)] != d {
		return errors.New("File in wrong directory")
	}
	return os.MkdirAll(sd, 0755)
}

func (s *Server) Save(p *Page, msg string) error {
	filename := p.Title + FileExtension
	path := path.Join(s.config.data, filename)

	if err := mkSubDir(s.config.data, filename); err != nil {
		log.Println("mkdir:", err)
		return err
	}

	if err := ioutil.WriteFile(path, p.Body, 0600); err != nil {
		log.Println("write file:", path)
		return err
	}

	if s.repo != nil {
		if err := s.repo.Save(filename, &commit{message: msg}, s.config.git.push); err != nil {
			log.Println("failed to save to repo:", filename)
			return err
		}
	}

	return nil
}

type ListFile struct {
	Info os.FileInfo
	Dir  string
}

func ListFiles(base, file string) []ListFile {
	dir := path.Join(base, file)
	//log.Println("list files in:", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	//log.Printf("found files: %+v\n", files)

	res := make([]ListFile, 0, len(files))
	for _, f := range files {
		res = append(res, ListFile{
			Info: f,
			Dir:  file,
		})
	}

	return res
}

// LoadPage ...
func LoadPage(title string, config Config, baseurl *url.URL) (*Page, error) {
	filename := path.Join(config.data, title+FileExtension)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	mtime := fi.ModTime()

	// Process and Parse the Markdown content
	// Also automatically replace CamelCase page identifiers as links
	markdown := AutoCamelCase(body, baseurl.String())

	unsafe := renderMarkdown(markdown)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

	return &Page{
		Title: title,
		Body:  body,
		HTML:  template.HTML(html),
		Brand: config.brand,
		Date:  mtime,
		Files: ListFiles(config.data, title),
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
	config    Config
	templates *Templates
	router    *httprouter.Router

	// Logger
	logger *logger.Logger

	// Stats/Metrics
	counters *Counters
	stats    *stats.Stats

	repo *Repo
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

		title := strings.TrimLeft(p.ByName("title"), "/")

		u, err := url.Parse("/view/")
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		baseurl := r.URL.ResolveReference(u)

		page, err := LoadPage(title, s.config, baseurl)
		if err != nil {
			page = &Page{Title: title, Brand: s.config.brand}
		}

		s.render("edit", w, page)
	}
}

// SaveHandler ...
func (s *Server) SaveHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_save")

		title := strings.TrimLeft(p.ByName("title"), "/")

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		// get body and sanitize newlines
		body := CleanNewlines(r.Form.Get("body"))

		msg := r.Form.Get("message")

		page := &Page{Title: title, Body: []byte(body), Brand: s.config.brand}
		err = s.Save(page, msg)
		if err != nil {
			log.Println("save page:", err)
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

		title := strings.TrimLeft(p.ByName("title"), "/")

		u, err := url.Parse("/view/")
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		baseurl := r.URL.ResolveReference(u)

		page, err := LoadPage(title, s.config, baseurl)
		if err != nil {
			u, err := url.Parse(fmt.Sprintf("/edit/%s", title))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			http.Redirect(
				w, r, r.URL.ResolveReference(u).String(), http.StatusFound,
			)

			return
		}

		s.render("view", w, page)
	}
}

var multipartByReader = &multipart.Form{
	Value: make(map[string][]string),
	File:  make(map[string][]*multipart.FileHeader),
}

func (s *Server) FileHandler() httprouter.Handle {
	var maxMemory int64 = 1024 * 1024 * 20
	type FormFile struct {
		File   multipart.File
		Header *multipart.FileHeader
	}

	GetFormFiles := func(r *http.Request, key string) ([]FormFile, error) {
		if r.MultipartForm == multipartByReader {
			return nil, errors.New("http: multipart handled by MultipartReader")
		}
		if r.MultipartForm == nil {
			err := r.ParseMultipartForm(maxMemory)
			if err != nil {
				return nil, err
			}
		}
		if r.MultipartForm != nil && r.MultipartForm.File != nil {
			if fhs := r.MultipartForm.File[key]; len(fhs) > 0 {
				files := make([]FormFile, 0)
				for _, fh := range fhs {
					f, err := fh.Open()
					if err != nil {
						return nil, err
					}
					files = append(files, FormFile{f, fh})
				}
				return files, nil
			}
		}
		return nil, http.ErrMissingFile
	}

	type UploadInfo struct {
		Name string
		Size int64
	}
	type UploadFile struct {
		Dir  string
		Info UploadInfo
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		title := strings.TrimLeft(p.ByName("title"), "/")
		filename := path.Join(s.config.data, title+FileExtension)
		dir := path.Join(s.config.data, title)

		log.Println("GETTING A FILE???", filename)

		files, err := GetFormFiles(r, "file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		results := make([]UploadFile, 0, len(files))
		for _, file := range files {
			defer file.File.Close()

			handler := file.Header

			log.Printf("Filename: %+v", handler.Filename)
			log.Printf("Size: %+v", handler.Size)
			log.Printf("Header: %+v", handler.Header)
			log.Printf("Dir: %+v", dir)

			if err := os.MkdirAll(dir, 0755); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpfile, err := ioutil.TempFile(dir, ".upload-*.tmp")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer tmpfile.Close()

			n, err := io.Copy(tmpfile, file.File)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if n != handler.Size {
				log.Printf("got less bytes than expected: %d < %d", handler.Size, n)
			}

			dstfile := path.Join(dir, handler.Filename)
			if err := os.Rename(tmpfile.Name(), dstfile); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			results = append(results, UploadFile{
				Dir: title,
				Info: UploadInfo{
					Name: handler.Filename,
					Size: handler.Size,
				},
			})
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bs, err := json.Marshal(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
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

// SearchHandler - handles searching for text in the wiki
func (s *Server) SearchHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := r.ParseForm(); err != nil {
			s.logger.Printf("ERROR: %s\n", err.Error())
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		bs, err := json.Marshal(s.DoSearch(r.FormValue("search")))
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
			s.config.bind,
			s.logger.Handler(
				s.stats.Handler(s.router),
			),
		),
	)
}

func (s *Server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.StatsHandler())

	s.router.ServeFiles("/css/*filepath", rice.MustFindBox("static/css").HTTPBox())
	s.router.ServeFiles("/js/*filepath", rice.MustFindBox("static/js").HTTPBox())
	s.router.ServeFiles("/f/*filepath", rice.MustFindBox("static/favicon").HTTPBox())
	fs := wikiFileSystem{http.Dir(s.config.data), s.config.data}
	s.router.ServeFiles("/file/*filepath", fs)

	s.router.GET("/", s.IndexHandler())
	s.router.GET("/view/*title", s.ViewHandler())
	s.router.GET("/edit/*title", s.EditHandler())
	s.router.POST("/file/*title", s.FileHandler())
	s.router.POST("/save/*title", s.SaveHandler())
	s.router.POST("/search", s.SearchHandler())
}

// NewServer ...
func NewServer(config Config) (*Server, error) {
	var repo *Repo

	if config.git.url != "" {
		r, err := newRepo(config.git.url, config.data)
		if err != nil {
			log.Println("newRepo:", config.data)
			return nil, err
		}
		repo = r
	}

	server := &Server{
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

		repo: repo,
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

	return server, nil
}
