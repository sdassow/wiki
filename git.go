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

func newRepo(repo string, dir string) (*Repo, error) {
	// open existing working copy
	r, err := git.PlainOpen(dir)
	if err != nil {
		if !errors.Is(err, git.ErrRepositoryNotExists) {
			log.Println("failed to open existing repository:", dir, err)
			return nil, err
		}

		// clone repo from url
		log.Println("cloning repository:", repo)

		r, err = git.PlainClone(dir, false, &git.CloneOptions{URL: repo})
		if err != nil {
			if err.Error() != "repository not found" {
				log.Println("failed to clone repository:", repo, ":", err)
				return nil, err
			}

			log.Println("initialize new local repository:", dir)

			// if it's local we can create a new repo
			r, err = git.PlainInit(dir, false)
			if err != nil {
				return nil, err
			}
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

	//log.Println("adding file:", filename)

	_, err := r.w.Add(filename)
	if err != nil {
		log.Println("add file:", filename, err)
		return err
	}
	//log.Println("blobHash:", blobHash)

	_, err = r.w.Status()
	if err != nil {
		log.Println("status:", err)
		return err
	}

	//log.Println("status:", status)

	_, err = r.w.Commit(co.message, &git.CommitOptions{
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

	//log.Println("commit hash:", hash)

	if push {
		if err := r.r.Push(&git.PushOptions{}); err != nil {
			if err.Error() != "remote not found" {
				log.Println("push:", err)
				return err
			}
		}
	}

	return nil
}
