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
  matches, _ := ctrl.templates(wc, "new.html")
  t := template.Must(template.ParseFiles(matches...))
  var output bytes.Buffer
  data := struct {
    Title string
  }{
    "Test Page",
  }
  t.Execute(&output, data )
  wc.Aec.Infof("Template: %v", t.Tree)
  return goweb.Respond.With(c, 200, output.Bytes())
}

func (ctrl *ContentController) templates(wc mycontext.Context, main string) ([]string, error) {
  pattern := filepath.Join("views", "partials", "*.html")
  wc.Aec.Infof("Pattern: %v", pattern)
  matches, err := filepath.Glob(pattern)
  if err != nil {
    wc.Aec.Errorf("Error finding files: %v", err)
  } else {
    matches = append([]string{filepath.Join("views", main)}, matches...)
    wc.Aec.Infof("Found matches: %v", matches)
  }
  return matches, err
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

  matches, _ := ctrl.templates(wc, "draft.html")
  t := template.Must(template.ParseFiles(matches...))
  var page bytes.Buffer
  draftdata := struct {
    Draft template.HTML
  }{
    template.HTML(output.String()),
  }
  t.Execute(&page, draftdata )

  return goweb.Respond.With(c, 200, page.Bytes())

}
