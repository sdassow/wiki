package main

import (
	"github.com/go-git/go-git/v5"
	//"github.com/go-git/go-git/storage/memory"
	"github.com/go-git/go-git/v5/plumbing/object"

	"errors"
	"log"
	"time"
	//"os"
	"os/user"
	"sync"
	//"strings"
)

type Repo struct {
	sync.Mutex
	r *git.Repository
	w *git.Worktree
}

type commit struct {
	name      string
	email     string
	timestamp time.Time
	message   string
}

func newRepo(url string, dir string) (*Repo, error) {
	// try to open existing repo
	r, err := git.PlainOpen(dir)
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryNotExists) {
			log.Println("open existing repo:", dir, err)
			return nil, err
		}
		log.Println("clone repo...", url)
		r, err = git.PlainClone(dir, false, &git.CloneOptions{URL: url})
		if err != nil {
			log.Println("clone repo:", dir, err)
			return nil, err
		}
	}
	w, err := r.Worktree()
	if err != nil {
		log.Println("get worktree:", err)
		return nil, err
	}
	return &Repo{
		r: r,
		w: w,
	}, nil
}

func getCommitDefaults(co *commit) *commit {
	if co == nil {
		co = &commit{}
	}
	if co.name == "" {
		u, err := user.Current()
		if err != nil {
			co.name = "Unknown User"
		} else {
			co.name = u.Name
		}
	}
	if co.email == "" {
		u, err := user.Current()
		if err != nil {
			co.email = "unknown@localhost"
		} else {
			co.email = u.Username + "@localhost"
		}
	}
	if co.timestamp.IsZero() {
		co.timestamp = time.Now()
	}
	if co.message == "" {
		co.message = "making a change"
	}

	return co
}

// file is on disk, now gets added to git
func (r *Repo) Save(filename string, co *commit, push bool) error {
	r.Lock()
	defer r.Unlock()

	co = getCommitDefaults(co)

	log.Println("adding file:", filename)

	blobHash, err := r.w.Add(filename)
	if err != nil {
		log.Println("add file:", filename, err)
		return err
	}
	log.Println("blobHash:", blobHash)

	status, err := r.w.Status()
	if err != nil {
		log.Println("status:", err)
		return err
	}

	log.Println("status:", status)

	hash, err := r.w.Commit(co.message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  co.name,
			Email: co.email,
			When:  co.timestamp,
		},
	})
	if err != nil {
		log.Println("commit:", err)
		return err
	}

	log.Println("hash:", hash)

	if push {
		if err := r.r.Push(&git.PushOptions{}); err != nil {
			log.Println("push:", err)
			return err
		}
	}

	return nil
}
