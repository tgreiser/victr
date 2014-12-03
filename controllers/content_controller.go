package controllers

import (
  "appengine"
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

/*
This controller is poorly named, but it manages both content and pages
Content are the versioned entities
Pages hold the current view and metadata
*/
type ContentController struct {
  BaseController
}

func (ctrl *ContentController) Read(key string, c context.Context) error {
  if key == "new" { return ctrl.New(c) }
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Content Read (browse versions)")

  msg := ""
  var p models.Page
  k := models.DsKey(wc, "Page", key)
  ckey := wc.Ctx.FormValue("content_key")
  var content models.Content

  if err := models.FindPage(wc, k, &p); err != nil {
    wc.Aec.Infof("Page not found: %v %v %v", k, key, err)
    return ctrl.error(wc, "Page not found", err)
  }

  if ckey != "" {
    k, err := datastore.DecodeKey(ckey)
    if err == nil {
      if err := models.FindContent(wc, k, &content); err != nil {
        wc.Aec.Infof("Unable to load content: %v %v", ckey, err)
      }
    }
  }
  if content.Key == nil {
    if err := models.FindContent(wc, p.CurrentVersionKey, &content); err != nil {
      wc.Aec.Infof("Unable to load current version content: %v %v", p.CurrentVersionKey, err)
    }
  }


  return ctrl.renderRead(wc, msg, map[string]string{}, &p, &content)
}

func (ctrl *ContentController) renderRead(wc mycontext.Context, message string, errs map[string]string, page *models.Page, content *models.Content) error {
  // load all the versions for this page
  ckey := wc.Ctx.FormValue("content_key")
  k := &datastore.Key{}
  var err error
  if ckey != "" {
    wc.Aec.Infof("Get content: %v %v", page.Key, wc.Ctx.FormValue("content_key"))
    k, err = datastore.DecodeKey(wc.Ctx.FormValue("content_key"))
    if err != nil { return ctrl.error(wc, "err_serious", err) }
    wc.Aec.Infof("Got key: %v", k)
  } else {
    k = page.CurrentVersionKey
  }
  versions, err := models.FetchContentByPage(wc, page.Key, k)
  if err != nil { return ctrl.error(wc, "err_serious", err) }
  sites, def_site, err := ctrl.prepSites(wc)
  if err != nil { return ctrl.error(wc, "err_serious", err) }
  themes, err := models.FetchThemes(wc, def_site.Bucket, def_site.Theme)
  if err != nil || len(themes) == 0 {
    return ctrl.error(wc, "err_no_themes", err)
  }

  data := struct {
    Message string
    Page *models.Page
    Content *models.Content
    Errors map[string]string
    Sites []*models.Site
    Themes []*models.Theme
    Versions []*models.Content
  } {
    message,
    page,
    content,
    errs,
    sites,
    themes,
    versions,
  }

  return ctrl.render(wc, "pageview", data)
}

func (ctrl *ContentController) ReadMany(c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Content ReadMany")
  return ctrl.renderReadMany(wc, "")
}

func (ctrl *ContentController) renderReadMany(wc mycontext.Context, msg string) error {
  limit := 100
  offset := 0

  sites, def_site, err := ctrl.prepSites(wc)
  if err != nil { return err }

  pages, err := models.FetchPages(wc, def_site.Key, limit, offset)
  if err != nil {
    msg := "Unable to find site contents!"
    ctrl.error(wc, msg, err)
  }

  return ctrl.render(wc, "content", struct {
    Pages []*models.Page
    Sites []*models.Site
    Message template.HTML
  } {
    pages,
    sites,
    template.HTML(msg),
  })
}

/*
Show the markdown editor form
*/
func (ctrl *ContentController) New(c context.Context) error {
  wc := mycontext.NewContext(c)
  return ctrl.renderNew(wc, "", map[string]string{}, nil, nil)
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

func (ctrl *ContentController) renderNew(wc mycontext.Context, message string, errs map[string]string, edit *models.Content, page *models.Page) error {
  sites, def_site, err := ctrl.prepSites(wc)
  if err != nil { return err }
  themes, err := models.FetchThemes(wc, def_site.Bucket, def_site.Theme)
  if err != nil || len(themes) == 0 {
    return ctrl.error(wc, "err_no_themes", err)
  }
  if edit == nil {
    edit = &models.Content{
      Markdown: `This is the *first* editor.
------------------------------

Just plain **Markdown**, except that the input is sanitized:

<marquee>I'm the ghost from the past!</marquee>`,
    }
  }
  wc.Aec.Infof("Got content: %v", edit)
  data := struct {
    Sites []*models.Site
    Errors map[string]string
    Themes []*models.Theme
    Message template.HTML
    Content *models.Content
    Page *models.Page
  } {
    sites,
    errs,
    themes,
    template.HTML(message),
    edit,
    page,
  }
//  wc.Aec.Infof("Page data: %v %v", edit.Title, page.Path)
  return ctrl.render(wc, "new", data)
}

func (ctrl *ContentController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Running create")

  content := &models.Content{}
  page := models.NewPage(wc)
  wc.Aec.Infof("Returned page: %v", page)
  errs := page.Validate(wc)
  if len(errs) > 0 {
    msg := "Failed to validate new page"
    wc.Aec.Warningf("%v: #%v %v", msg, len(errs), errs)
    return ctrl.renderNew(wc, msg, errs, models.NewContent(wc, page.Key), page)
  }

  err := datastore.RunInTransaction(wc.Aec, func(c appengine.Context) error {
    if err := page.Save(wc, page.Key); err != nil {
      return ctrl.renderNew(wc, "Failed to save page", map[string]string{}, nil, page)
    }

    content = models.NewContent(wc, page.Key)
    content.Draft = true
    wc.Aec.Infof("Content %v", content)
    errs2 := content.Validate()
    for k, v := range errs2 {
      errs[k] = v
    }

    wc.Aec.Infof("Saving draft...")
    content.Version = page.CurrentVersion
    if err := content.Save(wc, models.NewContentKey(wc)); err != nil {
      return ctrl.renderNew(wc, "Failed to save content", map[string]string{}, content, page)
    }

    page.CurrentVersionKey = content.Key
    if err := page.Save(wc, page.Key); err != nil {
      return ctrl.renderNew(wc, "Failed to save page", map[string]string{}, content, page)
    }

    return nil
  }, nil)
  if err != nil {
    msg := "Failed to save new page"
    wc.Aec.Errorf("%v: %v", msg, err)
    return ctrl.renderNew(wc, msg, map[string]string{}, content, page)
  }

  output := content.Build(wc)
  pagedata := struct {
    Draft template.HTML
    Title string
    Content string
    Path string
    Key string
    PageKey string
  } {
    template.HTML(output.String()),
    content.Title,
    content.Markdown,
    page.Path,
    content.Key.Encode(),
    models.NiceKey(page.Key),
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
    wc.Aec.Errorf("Failed to decode site key, can not publish: %v %v", wc.Ctx.FormValue("key"), err)
    return ctrl.renderNew(wc, "Could not publish", map[string]string{}, nil, nil)
  }
  var content models.Content
  if err := models.FindContent(wc, key, &content); err != nil {
    wc.Aec.Errorf("Failed to load content, could not publish: %v %v", key, err)
    return ctrl.renderNew(wc, "Could not publish", map[string]string{}, nil, nil)
  }
  var page models.Page
  if err := models.FindPage(wc, content.PageKey, &page); err != nil {
    wc.Aec.Errorf("Failed to load page, could not publish: %v %v", key, err)
    return ctrl.renderNew(wc, "Could not publish", map[string]string{}, &content, nil)
  }
  output := content.Build(wc)

  // load page data
  wc.Aec.Infof("Publishing %v", content.Title)
  var site models.Site;
  if err := models.FindSite(wc, page.SiteKey, &site); err != nil {
    wc.Aec.Errorf("Unable to find site: %v", err)
    return err
  }

  err = datastore.RunInTransaction(wc.Aec, func(c appengine.Context) error {
    // save to datastore
    content.Draft = false
    content.Version = page.CurrentVersion
    if err := content.Save(wc, content.Key); err != nil {
      return err
    }

    page.Published = true
    if err := page.Save(wc, page.Key); err != nil {
      return err
    }
    return nil
  }, nil)
  if err != nil {
    wc.Aec.Errorf("Failed to save page or content: %v", err)
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
    Name: page.Path,
  }

  object, err = obj.Insert(site.Bucket, object).Media(bytes.NewReader(output.Bytes())).Do()
  if err != nil {
    wc.Aec.Errorf("Failed to store page: %v", err)
    return ctrl.renderNew(wc, "Failed to upload published page!", map[string]string{}, &content, &page)
  }

  msg := "Page published at <a href=\"" + page.LiveUrl(wc) + "\" target=\"_blank\">" + page.LiveUrl(wc) +"</a>"
  return ctrl.renderReadMany(wc, msg)
}
