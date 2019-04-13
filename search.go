package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

const (
	bufferSize = 10
)

// SearchHit - an individual search match
type SearchHit struct {
	Title   string
	Page    string
	Subtext string
}

// SearchResults - container for results of searching
type SearchResults struct {
	Term  string
	Hits  []*SearchHit
	Error error
}

// DoSearch - searches the wiki for the given text
func (s *Server) DoSearch(term string) *SearchResults {
	results := &SearchResults{}
	results.Term = term
	results.Hits = make([]*SearchHit, 0)
	files, err := ioutil.ReadDir(s.config.data)
	if err != nil {
		s.logger.Printf("ERROR Reading Data Folder: %s\n", err.Error())
		results.Error = err
		return results
	}
	for _, f := range files {
		fullFile := filepath.Join(s.config.data, f.Name())
		if filepath.Ext(fullFile) == ".txt" {
			hit, err := checkFile(term, fullFile)
			if err != nil {
				s.logger.Printf("ERROR Reading File[%s]: %s\n", fullFile, err.Error())
			}
			if hit != nil {
				hit.Title = f.Name()
				hit.Title = hit.Title[:len(hit.Title)-4]
				hit.Page, _ = filepath.Rel(s.config.data, hit.Page)
				n := strings.IndexByte(hit.Page, '.')
				if n > 0 {
					hit.Page = "/view/" + hit.Page[:n] + "/"
				} else {
					hit.Page = "/view/" + hit.Page + "/"
				}
				results.Hits = append(results.Hits, hit)
			}
		}
	}
	return results
}

func checkFile(term string, filepath string) (*SearchHit, error) {

	rawData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	content := string(rawData)
	contentLength := len(content)

	matcher := search.New(language.English, search.Loose, search.IgnoreCase)
	pattern := matcher.CompileString(term)
	start, end := pattern.IndexString(content)
	if start > -1 {
		hit := &SearchHit{}
		hit.Page = filepath
		tStart := start - bufferSize
		if tStart < 0 {
			tStart = 0
		}
		tEnd := end + 10
		if tEnd > contentLength {
			tEnd = contentLength
		}
		hit.Subtext = content[tStart:tEnd]
		return hit, nil
	}

	return nil, nil
}
