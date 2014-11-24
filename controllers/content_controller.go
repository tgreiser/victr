package controllers

import (
  "bytes"
  "fmt"
  "html/template"
  "path"

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
    return goweb.Respond.WithRedirect(wc.Ctx, fmt.Sprintf("/sites/?msg=%s", wc.T("err_create_site")))
  }
  def_site := sites[0]
  wc.Aec.Infof("Def site theme: %v", def_site.Theme)
  themes, err := models.FetchThemes(wc, def_site.Theme)
  if err != nil || len(themes) == 0 {
    return ctrl.error(wc, "err_no_themes")
  }
  data := struct {
    Sites []*models.Site
    Errors map[string]string
    Themes []*models.Theme
  } {
    sites,
    map[string]string {},
    themes,
  }
  return ctrl.render(wc, "new", data)
}

func (ctrl *ContentController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)

  wc.Aec.Infof("Running create %v", c.FormValue("content"))

  tmpl := path.Join("themes", c.FormValue("theme"), "index.html")
  draft := template.Must(template.ParseFiles(tmpl))
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
