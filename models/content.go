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
    Version: 0,
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

type Content struct {
  Key *datastore.Key `datastore:"-"`
  SiteKey *datastore.Key
  Theme string
  Title string
  Path string
  Draft bool
  // if Draft=true, version is irrelevant
  Version int
  Markdown string `datastore:",noindex"`
  CreatedAt time.Time
  UpdatedAt time.Time
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
    c.UpdatedAt = time.Now()
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
  data := struct {
    Content template.HTML
    Title string
  }{
    template.HTML(blackfriday.MarkdownBasic([]byte(c.Markdown))),
    c.Title,
  }
  draft.Execute(&output, data )
  return output
}
