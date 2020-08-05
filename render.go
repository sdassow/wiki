package main

import (
	"github.com/russross/blackfriday/v2"
	"io"
)

type wikiRenderer struct {
	defR  *blackfriday.HTMLRenderer
	fileR *blackfriday.HTMLRenderer
}

func (r *wikiRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	switch node.Type {
	case blackfriday.Image:
		//node.LinkData.Destination = []byte(string(node.LinkData.Destination) + "#image")
		return r.fileR.RenderNode(w, node, entering)
	default:
		return r.defR.RenderNode(w, node, entering)
	}
}

func (r *wikiRenderer) RenderHeader(w io.Writer, node *blackfriday.Node) {
	r.defR.RenderHeader(w, node)
}

func (r *wikiRenderer) RenderFooter(w io.Writer, node *blackfriday.Node) {
	r.defR.RenderFooter(w, node)
}

func renderMarkdown(src []byte) []byte {
	r := wikiRenderer{
		defR: blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags,
			//AbsolutePrefix: "/view",
		}),
		fileR: blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags:          blackfriday.CommonHTMLFlags,
			AbsolutePrefix: "/file",
		}),
	}

	//out := blackfriday.Run(src, blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.NoEmptyLineBeforeBlock))
	out := blackfriday.Run(src, blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.NoEmptyLineBeforeBlock), blackfriday.WithRenderer(&r))

	return out
}
