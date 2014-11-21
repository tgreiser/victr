package controllers

import (
  "bytes"
  "fmt"
  "html/template"

  "github.com/russross/blackfriday"
  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"

  mycontext "github.com/tgreiser/victr/context"
  "github.com/tgreiser/victr/models"
)

type ContentController struct {
  BaseController
}

/*
Show the markdown editor form
*/
func (ctrl *ContentController) New(c context.Context) error {
  wc := mycontext.NewContext(c)
  sites, err := models.FetchSites(wc, 100, 0)
  if err != nil || len(sites) == 0 {
    return goweb.Respond.WithRedirect(wc.Ctx, fmt.Sprintf("/sites/?msg=%s", "Please create a site"))
  }
  data := struct {
    Sites []*models.Site
  } {
    sites,
  }
  return ctrl.render(wc, "new", data)
}

func (ctrl *ContentController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)

  wc.Aec.Infof("Running create %v", c.FormValue("content"))

  draft := template.Must(template.ParseFiles("themes/simple/index.html"))
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
