package controllers

import (
  "bytes"
  "encoding/base64"
  "fmt"
  "html/template"
  "net/http"
  "path"

  "code.google.com/p/google-api-go-client/storage/v1"
  "github.com/golang/oauth2"
  "github.com/golang/oauth2/google"
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

  output, title := ctrl.BuildContent(wc)

  pagedata := struct {
    Draft template.HTML
    Title string
    Content string
    Path string
  } {
    template.HTML(output.String()),
    title,
    c.FormValue("content"),
    c.FormValue("path"),
  }
  return ctrl.render(wc, "draft", pagedata)
}

func (ctrl *ContentController) BuildContent(wc mycontext.Context) (bytes.Buffer, string) {
  tmpl := path.Join("themes", wc.Ctx.FormValue("theme"), "index.html")
  draft := template.Must(template.ParseFiles(tmpl))
  var output bytes.Buffer
  data := struct {
    Content template.HTML
    Title string
  }{
    template.HTML(blackfriday.MarkdownBasic([]byte(wc.Ctx.FormValue("content")))),
    wc.Ctx.FormValue("title"),
  }
  draft.Execute(&output, data )
  return output, data.Title
}

func (ctrl *ContentController) contentData(wc mycontext.Context) {

}

/*
1. Save to datastore
2. Save rendered template to cloud storage
3. Redirect user to new URL
*/
func (ctrl *ContentController) Publish(c context.Context) error {
  wc := mycontext.NewContext(c)

  // load page data
  output, title := ctrl.BuildContent(wc)
  wc.Aec.Infof("Publishing %v", title)
  var site models.Site;
  if err := models.FindSiteFromEnc(wc, wc.Ctx.FormValue("site"), &site); err != nil {
    wc.Aec.Errorf("Unable to find site: %v", err)
    return err
  }

  // TODO save to datastore

  // get oauth client
  f, err := oauth2.New(
    google.AppEngineContext(wc.Aec),
    oauth2.Scope(
      "https://www.googleapis.com/auth/devstorage.read_write",
    ),
  )
  if err != nil {
    wc.Aec.Errorf("cloud storage auth failed: %v", err)
    // TODO return to /content/new with pagedata pre-filled
    return err
  }
  client := http.Client{Transport: f.NewTransport()}

  // do the cloud storage put operation
  storeSvc, err := storage.New(&client)
  if err != nil {
    wc.Aec.Errorf("failed to get storage client: %v", err)
    return err
  }
  obj := storage.NewObjectsService(storeSvc)
  object := &storage.Object {
    Bucket: site.Bucket,
    ContentType: "text/html",
    Name: wc.Ctx.FormValue("path"),
  }

  object, err = obj.Insert(site.Bucket, object).Media(base64.NewDecoder(base64.StdEncoding,&output)).Do()

  return goweb.Respond.WithRedirect(wc.Ctx, "/content/new")
}
