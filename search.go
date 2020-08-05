package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ego/riot/types"
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

func (s *Server) IndexPage(p *Page) {
	s.searcher.Index(p.Title+FileExtension, types.DocData{Content: string(p.Body)})
	s.searcher.Flush()
}

func (s *Server) SetupSearch() error {
	docs := s.searcher.NumDocsIndexed()

	log.Printf("index size: %+v", docs)

	if docs > 0 {
		return nil
	}

	log.Printf("scanning files: %+s", s.config.data)
	err := filepath.Walk(s.config.data, func(fullFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if filepath.Ext(fullFile) == FileExtension {
			rawData, err := ioutil.ReadFile(fullFile)
			if err != nil {
				return err
			}
			data := types.DocData{Content: string(rawData)}
			relFile, err := filepath.Rel(s.config.data, fullFile)
			if err != nil {
				return err
			}

			log.Printf("indexing %s", relFile)
			s.searcher.Index(relFile, data)
		}
		return nil
	})
	s.searcher.Flush()
	log.Println("indexing done")
	return err
}

func (s *Server) DoSearch(term string) *SearchResults {
	results := &SearchResults{}
	results.Term = term
	results.Hits = make([]*SearchHit, 0)

	sreq := types.SearchReq{Text: term}
	sres := s.searcher.SearchDoc(sreq)

	//log.Printf("results: %+v", sres)

	for _, doc := range sres.Docs {
		fullFile := doc.DocId
		title := strings.TrimSuffix(fullFile, FileExtension)
		hit := &SearchHit{
			Title: title,
			Page:  "/view/" + title,
		}
		results.Hits = append(results.Hits, hit)
	}

	return results
}
