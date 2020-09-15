package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/fcgi"
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

	"github.com/go-ego/riot"
	"github.com/gorilla/csrf"
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
	CSRF  template.HTML
	Prefix string
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

	s.IndexPage(p)

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
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	html := p.SanitizeBytes(unsafe)

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

	searcher *riot.Engine
}

func (s *Server) render(name string, w http.ResponseWriter, ctx interface{}) {
	buf, err := s.templates.Exec(name, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = buf.WriteTo(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// IndexHandler ...
func (s *Server) IndexHandler() httprouter.Handle {
	prefix := "."
	if s.config.prefix != "" {
		prefix = s.config.prefix
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_index")

		u, err := url.Parse(fmt.Sprintf("%s/view/FrontPage", prefix))
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

		page.CSRF = csrf.TemplateField(r)
		page.Prefix = s.config.prefix

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

		page.CSRF = csrf.TemplateField(r)
		page.Prefix = s.config.prefix

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

func (s *Server) Protect(h httprouter.Handle) http.Handler {
	protect := csrf.Protect(
		s.config.cookie.key,
		csrf.Secure(!s.config.cookie.insecure),
		csrf.Path(s.config.prefix + "/"),
	)

	return protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := httprouter.ParamsFromContext(r.Context())
		h(w, r, p)
	}))
}

type StripPrefix struct {
	Prefix string
}

func (sp StripPrefix) Handler(h http.Handler) http.Handler {
	return http.StripPrefix(sp.Prefix, h)
}

type RebindProtector struct {
	Hostnames	[]string
}

func (rp RebindProtector) Handler(h http.Handler) http.Handler {
	hostnames := make(map[string]bool)

	for _, hostname := range rp.Hostnames {
		k := strings.ToLower(hostname)
		hostnames[k] = true
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostname := strings.ToLower(strings.TrimRight(r.Host, ":0123456789"))

		if _, found := hostnames[hostname]; !found {
			log.Printf("Failed to find host: %s", hostname)
			http.Error(w, "Unknown hostname", http.StatusInternalServerError)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	var lsn net.Listener = nil
	var err error

	if s.config.listen.network != "stdio" {
		_ = os.Remove(s.config.listen.address)
		lsn, err = net.Listen(s.config.listen.network, s.config.listen.address)
		if err != nil {
			log.Fatal(err)
		}
	}

	stripped := StripPrefix{s.config.prefix}.Handler(s.logger.Handler(s.stats.Handler(s.router)))
	handler := RebindProtector{s.config.hosts}.Handler(stripped)

	if s.config.listen.protocol == "fcgi" {
		err = fcgi.Serve(lsn, handler)
	} else {
		if s.config.listen.network == "stdio" {
			log.Fatal("http over stdio not supported")
		}

		srv := &http.Server{
			Handler: handler,
		}

		if s.config.listen.protocol == "https" {
			if s.config.tls.certfile == "" && s.config.tls.keyfile == "" {
				log.Printf("Generating TLS config for %v", s.config.hosts)

				srv.TLSConfig, err = s.generateTLSConfig(s.config.hosts)
				if err != nil {
					log.Fatal("Failed to generate TLS config:", err)
				}
			} else {
				log.Printf("Loading TLS config from %v / %v", s.config.tls.certfile, s.config.tls.keyfile)
			}

			// use empty cert/key to avoid opening files and use the manual config instead
			err = srv.ServeTLS(lsn, s.config.tls.certfile, s.config.tls.keyfile)
		} else {
			err = srv.Serve(lsn)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.StatsHandler())

	s.router.ServeFiles("/css/*filepath", rice.MustFindBox("static/css").HTTPBox())
	s.router.ServeFiles("/js/*filepath", rice.MustFindBox("static/js").HTTPBox())
	s.router.ServeFiles("/webfonts/*filepath", rice.MustFindBox("static/webfonts").HTTPBox())
	s.router.ServeFiles("/f/*filepath", rice.MustFindBox("static/favicon").HTTPBox())
	fs := wikiFileSystem{http.Dir(s.config.data), s.config.data}
	s.router.ServeFiles("/file/*filepath", fs)

	s.router.Handler("GET", "/", s.Protect(s.IndexHandler()))
	s.router.Handler("GET", "/view/*title", s.Protect(s.ViewHandler()))
	s.router.Handler("GET", "/edit/*title", s.Protect(s.EditHandler()))
	s.router.Handler("POST", "/file/*title", s.Protect(s.FileHandler()))
	s.router.Handler("POST", "/save/*title", s.Protect(s.SaveHandler()))
	s.router.Handler("POST", "/search", s.Protect(s.SearchHandler()))
}

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

	if config.cookie.keyfile != "" {
		b, err := ioutil.ReadFile(config.cookie.keyfile)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if b == nil {
			log.Printf("Failed to find cookie keyfile, generating new one: %s", config.cookie.keyfile)
			b = make([]byte, 32)
			rand.Seed(time.Now().UnixNano())
			rand.Read(b)
			if err := ioutil.WriteFile(config.cookie.keyfile, b, 0600); err != nil {
				return nil, err
			}
		}
		log.Printf("Using cookie keyfile: %s", config.cookie.keyfile)
		config.cookie.key = b
	}

	server := &Server{
		config:    config,
		router:    httprouter.New(),
		templates: NewTemplates("base"),

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "wiking",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
			OutputFlags:          log.LstdFlags,
		}),

		// Stats/Metrics
		counters: NewCounters(),
		stats:    stats.New(),

		repo:     repo,
		searcher: riot.New("en", config.indexdir),
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
	if err := server.SetupSearch(); err != nil {
		log.Println(err)
	}

	return server, nil
}
