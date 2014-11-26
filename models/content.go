package models

import (
  "appengine"
  "appengine/datastore"
  "bytes"
  "html/template"
  "path"
  "time"

  "github.com/russross/blackfriday"

  mycontext "github.com/tgreiser/victr/context"
)

func NewContentKey(wc mycontext.Context) *datastore.Key {
  return datastore.NewIncompleteKey(wc.Aec, "Content", nil)
}

func NewContent(wc mycontext.Context) *Content {
  content := &Content {
    Theme: wc.Ctx.FormValue("theme"),
    Title: wc.Ctx.FormValue("title"),
    Path: wc.Ctx.FormValue("path"),
    Markdown: wc.Ctx.FormValue("content"),
  }

  sitekey, err := datastore.DecodeKey(wc.Ctx.FormValue("site_id"))
  if err != nil {
    wc.Aec.Warningf("Failed to decode site key: %v %v", wc.Ctx.FormValue("site_id"), err)
    return content
  }
  content.SiteKey = sitekey
  wc.Aec.Infof("Params: %v", wc.Ctx.FormParams())
  wc.Aec.Infof("New content - site key: %v %v", wc.Ctx.FormValue("site_id"), sitekey)
  return content
}

func FindContent(wc mycontext.Context, k *datastore.Key, c *Content) error {
  if err := datastore.Get(wc.Aec, k, c); err != nil {
    if err != datastore.ErrNoSuchEntity {
      wc.Aec.Errorf("datastore error with FindContent: %v", err)
    }
    return err
  }
  c.Key = k
  return nil
}

func FetchContent(wc mycontext.Context, site_key *datastore.Key, limit, offset int) ([]*Content, error) {
  q := datastore.NewQuery("Content").
    Project("SiteKey", "Path", "Title").Distinct().
    Order("Title").
    Limit(limit).
    Offset(offset)
  contents := make([]*Content, 0, limit)
  keys, err := q.GetAll(wc.Aec, &contents)
  if _, ok := err.(*datastore.ErrFieldMismatch); ok {
    wc.Aec.Infof("datastore missing field, ignoring: %v", err)
    err = nil
  } else if err != nil {
    wc.Aec.Errorf("got error instead of content list: %v", err)
    return nil, err
  }

  for i, k := range keys {
    contents[i].Key = k
    wc.Aec.Infof("Contents: %v", contents[i])
  }
  return contents, err
}

type Content struct {
  Key *datastore.Key `datastore:"-"`
  SiteKey *datastore.Key
  Theme string
  Title string
  Path string
  Draft bool
  // version is achieved with CreatedAt
  Markdown string `datastore:",noindex"`
  CreatedAt time.Time
}

func (c *Content) Validate() map[string]string {
  ret := map[string]string{}

  if c.Theme == "" { ret["theme"] = "Please select a theme to use" }
  if c.Title == "" { ret["title"] = "Please enter a title for your page" }
  if c.Path == "" { ret["path"] = "Please enter the relative path where your file will be published" }
  if c.Markdown == "" { ret["markdown"] = "Please enter some content." }
  if c.SiteKey == nil { ret["site"] = "Please select a site" }
  return ret
}

func (c *Content) Save(wc mycontext.Context, key *datastore.Key) error {
  err := datastore.RunInTransaction(wc.Aec, func(aec appengine.Context) error {
    if c.Key == nil { c.CreatedAt = time.Now() }
    key, e := datastore.Put(aec, key, c)
    if e != nil {
      return e
    }
    c.Key = key
    return nil
  }, nil)
  if err != nil {
    wc.Aec.Errorf("datastore write failed: %v", err)
  }
  return err
}

func (c *Content) Build(wc mycontext.Context) bytes.Buffer {
  wc.Aec.Infof("Got form vals: %v", c)
  tmpl := path.Join("themes", c.Theme, "index.html")
  draft := template.Must(template.ParseFiles(tmpl))
  var output bytes.Buffer
  path := "http://"
  if wc.Ctx.HttpRequest().TLS != nil { path = "https://" }
  path = path + appengine.DefaultVersionHostname(wc.Aec) + "/themes/" + c.Theme + "/"
  data := struct {
    Content template.HTML
    Title string
    ThemePath string
  }{
    template.HTML(blackfriday.MarkdownBasic([]byte(c.Markdown))),
    c.Title,
    path,
  }
  draft.Execute(&output, data )
  return output
}

func (c *Content) LiveUrl(wc mycontext.Context) string {
  var site Site
  if e := FindSite(wc, c.SiteKey, &site); e != nil {
    wc.Aec.Errorf("error building URL, no site: %v", e)
    return "#"
  }
  return site.URL + "/" + c.Path
}
