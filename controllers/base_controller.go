package controllers

import (
  "bytes"
  "html/template"
  "net/http"
  "path/filepath"

  "github.com/stretchr/goweb"

  mycontext "github.com/tgreiser/victr/context"
)

type BaseController struct {}

func (ctrl *BaseController) render(wc mycontext.Context, main string, data interface {}) error {
  matches, _ := ctrl.templates(wc, main)
  wc.Aec.Infof("Got matches for %v.html: %v", main, matches)
  t := template.New("temp").Funcs(ctrl.funcMap())
  t = template.Must(t.ParseFiles(matches...))
  var nav bytes.Buffer
  t.ExecuteTemplate(&nav, "nav-form", data )
  var form bytes.Buffer
  t.ExecuteTemplate(&form, "form", data)
  var page bytes.Buffer
  t.ExecuteTemplate(&page, "page", data)

  var output bytes.Buffer
  pagedata := struct {
    Request *http.Request
    Form template.HTML
    NavForm template.HTML
    Page template.HTML
  } {
    wc.Ctx.HttpRequest(),
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

func (ctrl *BaseController) funcMap() template.FuncMap {
  return template.FuncMap{
    "menu_link": ctrl.MenuLink,
  }
}

func (ctrl *BaseController) MenuLink(url, title, current_url string) template.HTML {
  ret := "<li><a href=\""+url+"\">"+title+"</a></li>"
  if len(url) <= len(current_url) && url == current_url[0:len(url)] {
    ret = "<li class=\"active\"><a href=\""+url+"\">"+title+"<span class=\"sr-only\">(current)</span></a></li>"
  }
  return template.HTML(ret)
}
