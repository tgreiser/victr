package controllers

import (
  "bytes"
  "html/template"
  "path/filepath"

  "github.com/stretchr/goweb"

  mycontext "github.com/tgreiser/victr/context"
)

type BaseController struct {}

func (ctrl *BaseController) render(wc mycontext.Context, main string, data interface {}) error {
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

func (ctrl *BaseController) templates(wc mycontext.Context, main string) ([]string, error) {
  var matches [2]string
  matches[0] = mycontext.AppPath(filepath.Join("views", main + ".html"))
  matches[1] = mycontext.AppPath(filepath.Join("views", "base.html"))
  return matches[0:], nil
}
