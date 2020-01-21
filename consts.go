package main

import (
	"crypto/sha512"
	"encoding/base64"
)

func init() {
	sha := sha512.New()
	sha.Write([]byte(css))
	integrity = base64.StdEncoding.EncodeToString(sha.Sum(nil))
}

var integrity string

const css = `body {
	background-color: #fff;
	color: #000;
	font-family: monospace;
	font-size: 14px;
}

td, th {
	padding: 0 0.5em;
}

th {
	text-align: left;
}

tr:hover {
	background-color: #eee;
}

.num {
	text-align: right;
}

.desc {
	color: #444;
}
`

const layoutTmpl = `<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<meta name="viewport"
			content="width=device-width, initial-scale=1, shrink-to-fit=no">
		{{if .Repo }}{{if .Repo.Description}}
			<meta name="description" content="{{index .Repo.Description 0}}">
		{{end}}{{end}}
		<link rel="stylesheet" type="text/css" href="/style.css"
			integrity="sha512-{{.Integrity}}">
		<title>{{.Title}}</title>
	</head>
	<body>
		{{if .Repo}}
			<p><b>{{.Repo.Name}}</b>{{if .Repo.Bare}} ({{.Repo.Git.Ref}}, bare repository){{end}}</p>
			{{range .Repo.Description}}
				<p>{{.}}</p>
			{{end}}
			<p><a href="/{{.Repo.Name}}">Log</a>
				| <a href="/{{.Repo.Name}}/files">Files</a>
				| <a href="/">&lt;&lt; Repositories</a>{{else}}
			<p><b>{{.Title}}</b></p>
		{{end}}
		<hr>
		{{template "content" .}}
	</body>
</html>`

const reposTmpl = `{{define "content"}}<table>
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
			<td>{{if .Description}}{{index .Description 0}}{{end}}</td>
			<td>{{.Git.Ref}}{{if .Bare}} (bare){{end}}</td>
		</tr>
	{{end}}</tbody>
</table>{{end}}`

const logTmpl = `{{define "content"}}<table>
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

const lsTmpl = `{{define "content"}}<table>
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

const commitTmpl = `{{define "content"}}<pre>{{ printf "%s" .Commit.CatFile }}</pre>
	<hr>
	<pre>{{ printf "%s" .Commit.DiffStat }}</pre>
	<hr>
	<pre>{{ printf "%s" .Commit.Diff }}</pre>{{end}}`

const showTmpl = `{{define "content"}}{{if .Repo.Bare}}
	<p><b>(Cannot view files of bare repositories)</b></p>
	{{else if .Binary}}
	<p><b>(Binary file)</b></p>{{else}}<pre>{{ printf "%s" .File}}</pre>{{end}}{{end}}`
