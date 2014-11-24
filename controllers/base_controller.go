package controllers

import (
  "bytes"
  "fmt"
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
  if e := t.ExecuteTemplate(&output, "base", pagedata); e != nil {
    // unfortunately, this doesn't seem to fire if the template crashes mid-render due to a func error
    wc.Aec.Errorf("Failed to render: %v", e)
    return ctrl.error(wc, "err_serious")
  }

  return goweb.Respond.With(wc.Ctx, 200, output.Bytes())
}

func (ctrl *BaseController) error(wc mycontext.Context, msg_id string) error {
  data := struct {
    Message string
  } {
    wc.T(msg_id),
  }
  return ctrl.render(wc, "error", data)
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
    "fg": ctrl.FormGroup,
    "fg_close": ctrl.FormGroupClose,
  }
}

func (ctrl *BaseController) MenuLink(url, title, current_url string) template.HTML {
  ret := "<li><a href=\""+url+"\">"+title+"</a></li>"
  if len(url) <= len(current_url) && url == current_url[0:len(url)] {
    ret = "<li class=\"active\"><a href=\""+url+"\">"+title+"<span class=\"sr-only\">(current)</span></a></li>"
  }
  return template.HTML(ret)
}

func (ctrl *BaseController) FormGroup(name string, errs map[string]string, label string) template.HTML {
  cls := "form-group"
  if _, ok := errs[name]; ok {
    cls = cls + " alert alert-danger"
  }
  return template.HTML(fmt.Sprintf(`<div class="%s">
      <label for=\"%s\">%s:</label>`, cls, name, label))
}
func (ctrl *BaseController) FormGroupClose(name string, errs map[string]string) template.HTML {
  ret := "</div>"
  if err, ok := errs[name]; ok {
    ret = fmt.Sprintf("<span>%s</span>\n", err) + ret
  }
  return template.HTML(ret)
}
