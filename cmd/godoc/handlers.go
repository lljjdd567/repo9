// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The /doc/codewalk/ tree is synthesized from codewalk descriptions,
// files named $GOROOT/doc/codewalk/*.xml.
// For an example and a description of the format, see
// http://golang.org/doc/codewalk/codewalk or run godoc -http=:6060
// and see http://localhost:6060/doc/codewalk/codewalk .
// That page is itself a codewalk; the source code for it is
// $GOROOT/doc/codewalk/codewalk.xml.

package main

import (
	"log"
	"net/http"
	"regexp"
	"text/template"

	"code.google.com/p/go.tools/godoc"
	"code.google.com/p/go.tools/godoc/vfs"
)

var (
	pres *godoc.Presentation
	fs   = vfs.NameSpace{}
)

func registerHandlers(pres *godoc.Presentation) {
	if pres == nil {
		panic("nil Presentation")
	}
	http.HandleFunc("/doc/codewalk/", codewalk)
	http.Handle("/doc/play/", pres.FileServer())
	http.Handle("/robots.txt", pres.FileServer())
	http.Handle("/", pres)
	handlePathRedirects(pkgRedirects, "/pkg/")
	handlePathRedirects(cmdRedirects, "/cmd/")
	for prefix, redirect := range prefixHelpers {
		p := "/" + prefix + "/"
		h := makePrefixRedirectHandler(p, redirect)
		http.HandleFunc(p, h)
	}
	for path, redirect := range redirects {
		h := makeRedirectHandler(redirect)
		http.HandleFunc(path, h)
	}
}

func readTemplate(name string) *template.Template {
	if pres == nil {
		panic("no global Presentation set yet")
	}
	path := "lib/godoc/" + name

	// use underlying file system fs to read the template file
	// (cannot use template ParseFile functions directly)
	data, err := vfs.ReadFile(fs, path)
	if err != nil {
		log.Fatal("readTemplate: ", err)
	}
	// be explicit with errors (for app engine use)
	t, err := template.New(name).Funcs(pres.FuncMap()).Parse(string(data))
	if err != nil {
		log.Fatal("readTemplate: ", err)
	}
	return t
}

func readTemplates(p *godoc.Presentation, html bool) {
	p.PackageText = readTemplate("package.txt")
	p.SearchText = readTemplate("search.txt")

	if html {
		codewalkHTML = readTemplate("codewalk.html")
		codewalkdirHTML = readTemplate("codewalkdir.html")
		p.DirlistHTML = readTemplate("dirlist.html")
		p.ErrorHTML = readTemplate("error.html")
		p.ExampleHTML = readTemplate("example.html")
		p.GodocHTML = readTemplate("godoc.html")
		p.PackageHTML = readTemplate("package.html")
		p.SearchHTML = readTemplate("search.html")
		p.SearchDescXML = readTemplate("opensearch.xml")
	}
}

// Packages that were renamed between r60 and go1.
var pkgRedirects = map[string]string{
	"asn1":              "encoding/asn1",
	"big":               "math/big",
	"cmath":             "math/cmplx",
	"csv":               "encoding/csv",
	"exec":              "os/exec",
	"exp/template/html": "html/template",
	"gob":               "encoding/gob",
	"http":              "net/http",
	"http/cgi":          "net/http/cgi",
	"http/fcgi":         "net/http/fcgi",
	"http/httptest":     "net/http/httptest",
	"http/pprof":        "net/http/pprof",
	"json":              "encoding/json",
	"mail":              "net/mail",
	"rand":              "math/rand",
	"rpc":               "net/rpc",
	"rpc/jsonrpc":       "net/rpc/jsonrpc",
	"scanner":           "text/scanner",
	"smtp":              "net/smtp",
	"tabwriter":         "text/tabwriter",
	"template":          "text/template",
	"template/parse":    "text/template/parse",
	"url":               "net/url",
	"utf16":             "unicode/utf16",
	"utf8":              "unicode/utf8",
	"xml":               "encoding/xml",
}

// Commands that were renamed between r60 and go1.
var cmdRedirects = map[string]string{
	"gofix":     "fix",
	"goinstall": "go",
	"gopack":    "pack",
	"govet":     "vet",
	"goyacc":    "yacc",
}

var redirects = map[string]string{
	"/blog":   "http://blog.golang.org",
	"/build":  "http://build.golang.org",
	"/change": "https://code.google.com/p/go/source/list",
	"/cl":     "https://gocodereview.appspot.com/",
	"/doc/go_for_cpp_programmers.html": "https://code.google.com/p/go-wiki/wiki/GoForCPPProgrammers",
	"/doc/go_tutorial.html":            "http://tour.golang.org/",
	"/issue":                           "https://code.google.com/p/go/issues",
	"/issue/new":                       "https://code.google.com/p/go/issues/entry",
	"/issues":                          "https://code.google.com/p/go/issues",
	"/play":                            "http://play.golang.org",
	"/talks":                           "http://talks.golang.org",
	"/tour":                            "http://tour.golang.org",
	"/wiki":                            "https://code.google.com/p/go-wiki/w/list",
}

var prefixHelpers = map[string]string{
	"blog":   "http://blog.golang.org/",
	"change": "https://code.google.com/p/go/source/detail?r=",
	"cl":     "https://codereview.appspot.com/",
	"issue":  "https://code.google.com/p/go/issues/detail?id=",
	"play":   "http://play.golang.org/",
	"talks":  "http://talks.golang.org/",
	"wiki":   "https://code.google.com/p/go-wiki/wiki/",
}

func handlePathRedirects(redirects map[string]string, prefix string) {
	for source, target := range pkgRedirects {
		h := makeRedirectHandler(prefix + target + "/")
		p := prefix + source
		http.HandleFunc(p, h)
		http.HandleFunc(p+"/", h)
	}
}

func makeRedirectHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	}
}

var validId = regexp.MustCompile(`^[A-Za-z0-9-]*$`)

func makePrefixRedirectHandler(prefix, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if p := r.URL.Path; p == prefix {
			// redirect /prefix/ to /prefix
			http.Redirect(w, r, p[:len(p)-1], http.StatusFound)
			return
		}
		id := r.URL.Path[len(prefix):]
		if !validId.MatchString(id) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		target := baseURL + id
		http.Redirect(w, r, target, http.StatusFound)
	}
}
