package controllers

import (
  "bytes"
  "html/template"

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
  t := template.Must(template.ParseFiles("views/new.html"))
  var output bytes.Buffer
  data := struct {
    Title string
  }{
    "Test Page",
  }
  t.Execute(&output, data )
  return goweb.Respond.With(c, 200, output.Bytes())
}

func (ctrl *ContentController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)

  wc.Aec.Infof("Running create %v", c.FormValue("content"))

  t := template.Must(template.ParseFiles("templates/simple/index.html"))
  var output bytes.Buffer
  data := struct {
    Content template.HTML
    Title string
  }{
    template.HTML(blackfriday.MarkdownBasic([]byte(c.FormValue("content")))),
    c.FormValue("title"),
  }
  t.Execute(&output, data )
  return goweb.Respond.With(c, 200, output.Bytes())

}
