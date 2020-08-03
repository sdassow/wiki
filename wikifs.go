package main

import (
	"net/http"
	"os"
	"log"
	"path"
)

func (fs wikiFileSystem) isWikiFile(name string) bool {
	fullpath := path.Join(fs.basedir, name)
	basepath := path.Dir(fullpath) + FileExtension
	log.Printf("basepath: %s", basepath)
	_, err := os.Stat(basepath)
	if err != nil {
		return false
	}
	return true
}

type wikiFile struct {
	http.File
}

func (f wikiFile) Readdir(n int) (fis []os.FileInfo, err error) {
	files, err := f.File.Readdir(n)
	for _, file := range files {
		//if !isWikiFile(file.Name()) {
			fis = append(fis, file)
		//}
	}
	return
}

type wikiFileSystem struct {
	http.FileSystem
	basedir string
}

func (fs wikiFileSystem) Open(name string) (http.File, error) {
	if !fs.isWikiFile(name) {
		return nil, os.ErrPermission
	}
	file, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return wikiFile{file}, err
}
