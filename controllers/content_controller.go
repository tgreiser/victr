package controllers

import (
  "appengine/datastore"
  "bytes"
  "fmt"
  "html/template"
  "net/http"

  "code.google.com/p/google-api-go-client/storage/v1"
  "github.com/golang/oauth2"
  "github.com/golang/oauth2/google"
  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"

  mycontext "github.com/tgreiser/victr/context"
  "github.com/tgreiser/victr/models"
)

type ContentController struct {
  BaseController
}

func (ctrl *ContentController) ReadMany(c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Content ReadMany")

  limit := 100
  offset := 0

  sites, def_site, err := ctrl.prepSites(wc)
  if err != nil { return err }

  contents, err := models.FetchContent(wc, def_site.Key, limit, offset)
  if err != nil {
    msg := "Unable to find site contents!"
    ctrl.error(wc, msg)
    wc.Aec.Errorf("%v: %v", msg, err)
  }

  return ctrl.render(wc, "content", struct {
    Contents []*models.Content
    Sites []*models.Site
    Message string
  } {
    contents,
    sites,
    "",
  })
}

/*
Show the markdown editor form
*/
func (ctrl *ContentController) New(c context.Context) error {
  wc := mycontext.NewContext(c)
  return ctrl.renderNew(wc, "", map[string]string{}, nil)
}

func (ctrl* ContentController) prepSites(wc mycontext.Context) ([]*models.Site, *models.Site, error) {
  sites, err := models.FetchSites(wc, 100, 0)
  if err != nil || len(sites) == 0 {
    return nil, nil, goweb.Respond.WithRedirect(wc.Ctx, fmt.Sprintf("/sites/?msg=%s", wc.T("err_create_site")))
  }
  def_site := sites[0]
  wc.Aec.Infof("Def site theme: %v", def_site.Theme)

  return sites, def_site, nil
}

func (ctrl *ContentController) renderNew(wc mycontext.Context, message string, errs map[string]string, edit *models.Content) error {
  sites, def_site, err := ctrl.prepSites(wc)
  if err != nil { return err }
  themes, err := models.FetchThemes(wc, def_site.Theme)
  if err != nil || len(themes) == 0 {
    return ctrl.error(wc, "err_no_themes")
  }
  data := struct {
    Sites []*models.Site
    Errors map[string]string
    Themes []*models.Theme
    Message template.HTML
  } {
    sites,
    errs,
    themes,
    template.HTML(message),
  }
  return ctrl.render(wc, "new", data)
}

func (ctrl *ContentController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)

  wc.Aec.Infof("Running create %v", c.FormValue("content"))

  content := models.NewContent(wc)
  content.Draft = true

  if errs := content.Validate(); len(errs) > 0 {
    msg := "Failed to validate content"
    wc.Aec.Warningf("%v: #%v %v", msg, len(errs), errs)
    return ctrl.renderNew(wc, msg, errs, content)
  }

  wc.Aec.Infof("Saving draft...")
  if err := content.Save(wc, models.NewContentKey(wc)); err != nil {
    return ctrl.renderNew(wc, "Failed to save content", map[string]string{}, content)
  }

  output := content.Build(wc)
  pagedata := struct {
    Draft template.HTML
    Title string
    Content string
    Path string
    Key string
  } {
    template.HTML(output.String()),
    content.Title,
    content.Markdown,
    content.Path,
    content.Key.Encode(),
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

  key, err := datastore.DecodeKey(wc.Ctx.FormValue("key"))
  if err != nil {
    wc.Aec.Errorf("Failed to decode site key, can not publish: %v %v", wc.Ctx.FormValue("site_id"), err)
    return ctrl.renderNew(wc, "Could not publish", map[string]string{}, nil)
  }
  var content models.Content
  if err := models.FindContent(wc, key, &content); err != nil {
    wc.Aec.Errorf("Failed to load content, could not publish: %v %v", wc.Ctx.FormValue("site_id"), err)
    return ctrl.renderNew(wc, "Could not publish", map[string]string{}, nil)
  }
  output := content.Build(wc)

  // load page data
  wc.Aec.Infof("Publishing %v", content.Title)
  var site models.Site;
  if err := models.FindSite(wc, content.SiteKey, &site); err != nil {
    wc.Aec.Errorf("Unable to find site: %v", err)
    return err
  }

  // save to datastore
  content.Draft = false
  if err := content.Save(wc, content.Key); err != nil {
    return err
  }

  // get oauth client
  wc.Aec.Infof("Getting oauth client")
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
  wc.Aec.Infof("Cloud storage put...")
  storeSvc, err := storage.New(&client)
  if err != nil {
    wc.Aec.Errorf("failed to get storage client: %v", err)
    return err
  }
  obj := storage.NewObjectsService(storeSvc)
  object := &storage.Object {
    Bucket: site.Bucket,
    ContentType: "text/html",
    Name: content.Path,
  }

  object, err = obj.Insert(site.Bucket, object).Media(bytes.NewReader(output.Bytes())).Do()
  if err != nil {
    wc.Aec.Errorf("Failed to store page: %v", err)
    return ctrl.renderNew(wc, "Failed to upload published page!", map[string]string{}, &content)
  }

  msg := "Page published at <a href=\"" + content.LiveUrl(wc) + "\" target=\"_blank\">" + content.LiveUrl(wc) +"</a>"
  return ctrl.renderNew(wc, msg, map[string]string{}, nil)
}
