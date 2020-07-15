package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"os"

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

	err := filepath.Walk(s.config.data, func(fullFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if filepath.Ext(fullFile) == FileExtension {
			hit, err := checkFile(term, fullFile)
			if err != nil {
				s.logger.Printf("ERROR Reading File[%s]: %s\n", fullFile, err.Error())
			}
			if hit != nil {
				relFile, _ := filepath.Rel(s.config.data, fullFile)
				hit.Title = strings.TrimSuffix(relFile, FileExtension)
				hit.Page = "/view/" + hit.Title
				results.Hits = append(results.Hits, hit)
			}
		}

		return nil
	} )
	if err != nil {
		s.logger.Printf("ERROR Walking Directory[%s]: %s\n", s.config.data, err.Error())
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
