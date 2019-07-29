package main

import (
	"bufio"
	"bytes"
	"html/template"
)

type page struct {
	Repo  *repository
	Title string
}

const tmplLayout = `<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta name="viewport"
			content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{if .Repo}}<meta name="description" content="{{.Repo.Description}}">{{end}}
		<link rel="stylesheet" type="text/css" href="/style.css" integrity="sha512-W9Z7ENHmGOifqH2GDCG2fNFR36qLVMqkRwX4Zvn88dau35nmg84GbTTbVMfnaLo5sBGb8OdQ8kYNQBF4gwo1xQ==">
		<title>{{.Title}}</title>
	</head>
	<body>
		{{if .Repo}}<p><b>{{.Repo.Name}}</b>{{if .Repo.Bare}} ({{.Repo.Git.Ref}}, bare repository){{end}}</p>
		<p class="desc">{{.Repo.Description}}</p>
		<p><a href="/{{.Repo.Name}}">Log</a>
			| <a href="/{{.Repo.Name}}/files">Files</a>
			| <a href="/">&lt;&lt; Repositories</a>{{else}}
		<p><b>{{.Title}}</b></p>{{end}}
		<hr>
		{{template "content" .}}
	</body>
</html>`

const tmplRepos = `{{define "content"}}<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
			<th>Ref</th>
		</tr>
	</thead>
	<tbody>{{range .Repos}}
		<tr>
			<td><a href="/{{.Name}}">{{.Name}}</a></td>
			<td>{{.Description}}</td>
			<td>{{.Git.Ref}}{{if .Bare}} (bare){{end}}</td>
		</tr>
	{{end}}</tbody>
</table>{{end}}`

const tmplLog = `{{define "content"}}<table>
	<thead>
		<tr>
			<th>Date</th>
			<th>Commit Message</th>
			<th>Author</th>
			<th class="num">Files</th>
			<th class="num">+</th>
			<th class="num">-</th>
		</tr>
	</thead>
	<tbody>{{range .Items}}
		<tr>
			<td>{{.Time.UTC.Format "2006-01-02 15:04"}}</td>
			<td><a href="/{{$.Repo.Name}}/commit/{{.Hash}}">{{.Subject}}</a></td>
			<td>{{.Name}}</td>
			<td class="num">{{.Stat.Changed}}</td>
			<td class="num">{{.Stat.Insertions}}</td>
			<td class="num">{{.Stat.Deletions}}</td>
		</tr>
	{{end}}</tbody>
</table>{{end}}`

const tmplLs = `{{define "content"}}<table>
	<thead>
		<tr>
			<th>Mode</th>
			<th>Name</th>
			<th class="num">Size</th>
		</tr>
	</thead>
	<tbody>{{range .Items}}
		<tr>
			<td>{{.Mode}}</td>
			{{if $.Repo.Bare}}<td>{{.Name}}</td>{{else}}<td><a href="/{{$.Repo.Name}}/file/{{.Name}}">{{.Name}}</a></td>{{end}}
			<td class="num">{{.Size}}</td>
		</tr>
	{{end}}</tbody>
</table>{{end}}`

const tmplCommit = `{{define "content"}}<pre>{{ printf "%s" .Commit.CatFile }}</pre>
	<hr>
	<pre>{{ printf "%s" .Commit.DiffStat }}</pre>
	<hr>
	<pre>{{ printf "%s" .Commit.Diff }}</pre>{{end}}`

const tmplShow = `{{define "content"}}{{if .Repo.Bare}}
	<p><b>(Cannot view files of bare repositories)</b></p>
	{{else if .Binary}}
	<p><b>(Binary file)</b></p>{{else}}<pre>{{ printf "%s" .File}}</pre>{{end}}{{end}}`

var templates map[string]*template.Template

func initializeTemplates() (err error) {
	var tmpls = []struct {
		name, format string
	}{
		{"commit", tmplCommit},
		{"log", tmplLog},
		{"ls", tmplLs},
		{"show", tmplShow},
	}

	templates = make(map[string]*template.Template, len(tmpls))

	for _, tmpl := range tmpls {
		templates[tmpl.name], err = template.New(tmpl.name).Parse(tmpl.format)

		if err != nil {
			return
		}

		templates[tmpl.name], err = templates[tmpl.name].Parse(tmplLayout)

		if err != nil {
			return
		}
	}

	return
}

var index []byte

func initializeIndex() error {
	var page = struct {
		page
		Repos map[string]*repository
	}{
		page: page{
			Repo:  nil,
			Title: "Repositories",
		},
		Repos: repos,
	}

	tmpl, err := template.New("repos").Parse(tmplRepos)

	if err != nil {
		return err
	}

	tmpl, err = tmpl.Parse(tmplLayout)

	if err != nil {
		return err
	}

	var b bytes.Buffer

	w := bufio.NewWriter(&b)

	if err := tmpl.Execute(w, page); err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}

	index = b.Bytes()

	return nil
}
