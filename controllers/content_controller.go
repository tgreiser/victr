package controllers

import (
  "bytes"
  "html/template"
  "path/filepath"

  "github.com/russross/blackfriday"
  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"

  mycontext "github.com/tgreiser/victr/context"
)

type ContentController struct {}

/*
Show the markdown editor form
*/
func (ctrl *ContentController) New(c context.Context) error {
  wc := mycontext.NewContext(c)
  return ctrl.render(wc, "new", "")
}

func (ctrl *ContentController) templates(wc mycontext.Context, main string) ([]string, error) {
  var matches [2]string
  matches[0] = filepath.Join("views", main + ".html")
  matches[1] = filepath.Join("views", "base.html")
  return matches[0:], nil
}

func (ctrl *ContentController) render(wc mycontext.Context, main string, data interface {}) error {
  matches, _ := ctrl.templates(wc, main)
  wc.Aec.Infof("Got matches for %v.html: %v", main, matches)
  t := template.Must(template.ParseFiles(matches...))
  var nav bytes.Buffer
  t.ExecuteTemplate(&nav, "nav-form", data )
  var form bytes.Buffer
  t.ExecuteTemplate(&form, "form", data)
  var page bytes.Buffer
  t.ExecuteTemplate(&page, "page", data)

  var output bytes.Buffer
  pagedata := struct {
    Form template.HTML
    NavForm template.HTML
    Page template.HTML
  } {
    template.HTML(form.String()),
    template.HTML(nav.String()),
    template.HTML(page.String()),
  }
  wc.Aec.Infof("Got nav: %v", pagedata.NavForm)
  t.ExecuteTemplate(&output, "base", pagedata)

  return goweb.Respond.With(wc.Ctx, 200, output.Bytes())
}

func (ctrl *ContentController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)

  wc.Aec.Infof("Running create %v", c.FormValue("content"))

  draft := template.Must(template.ParseFiles("templates/simple/index.html"))
  var output bytes.Buffer
  data := struct {
    Content template.HTML
    Title string
  }{
    template.HTML(blackfriday.MarkdownBasic([]byte(c.FormValue("content")))),
    c.FormValue("title"),
  }
  draft.Execute(&output, data )

  pagedata := struct {
    Draft template.HTML
    Title string
    Content string
    Path string
  } {
    template.HTML(output.String()),
    data.Title,
    c.FormValue("content"),
    c.FormValue("path"),
  }
  return ctrl.render(wc, "draft", pagedata)
}

/*
1. Save to datastore
2. Save rendered template to cloud storage
3. Redirect user to new URL
*/
func (ctrl *ContentController) Publish(c context.Context) error {
  wc := mycontext.NewContext(c)
  // s
  var output bytes.Buffer
  return goweb.Respond.With(wc.Ctx, 200, output.Bytes())
}
