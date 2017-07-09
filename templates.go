package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"sync"
)

type TemplateMap map[string]*template.Template

type Templates struct {
	sync.Mutex

	base      string
	templates TemplateMap
}

func NewTemplates(base string) *Templates {
	return &Templates{
		base:      base,
		templates: make(TemplateMap),
	}
}

func (t *Templates) Add(name string, template *template.Template) {
	t.Lock()
	defer t.Unlock()

	t.templates[name] = template
}

func (t *Templates) Exec(name string, ctx interface{}) (io.WriterTo, error) {
	t.Lock()
	defer t.Unlock()

	template, ok := t.templates[name]
	if !ok {
		return nil, fmt.Errorf("no such template: %s", name)
	}

	buf := bytes.NewBuffer([]byte{})
	err := template.ExecuteTemplate(buf, t.base, ctx)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
