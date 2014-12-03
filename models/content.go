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

func NewContent(wc mycontext.Context, page_key *datastore.Key) *Content {
  content := &Content {
    Theme: wc.Ctx.FormValue("theme"),
    Title: wc.Ctx.FormValue("title"),
    Markdown: wc.Ctx.FormValue("content"),
    PageKey: page_key,
    Version: 1,
  }

  return content
}

func FindContent(wc mycontext.Context, k *datastore.Key, c *Content) error {
  if err := datastore.Get(wc.Aec, k, c); err != nil {
    if err != datastore.ErrNoSuchEntity {
      wc.Aec.Errorf("datastore error with FindContent: %v %v", k, err)
    }
    return err
  }
  c.Key = k
  wc.Aec.Infof("Loaded content: %v", c)
  return nil
}

func FetchContentByPage(wc mycontext.Context, page_key *datastore.Key, sel *datastore.Key) ([]*Content, error) {
  lmt := 50
  wc.Aec.Infof("Page_key filter %v", page_key)
  q := datastore.NewQuery("Content").Filter("PageKey=", page_key).Order("-Version").Limit(lmt)
  versions := make([]*Content, 0, lmt)
  keys, err := q.GetAll(wc.Aec, &versions)
  if _, ok := err.(*datastore.ErrFieldMismatch); ok {
    wc.Aec.Infof("datastore missing field, ignoring: %v", err)
    err = nil
  } else if err != nil {
    wc.Aec.Errorf("got error instead of contents: %v", err)
    return nil, err
  } else if len(keys) > 0 {
    for i, k := range keys {
      versions[i].Key = k
      wc.Aec.Infof("Comparing %v and %v", k, sel)
      if versions[i].Key.String() == sel.String() {
        wc.Aec.Infof("Selected version: %v", versions[i].Version)
        versions[i].Selected = true
      }
    }
    return versions, nil
  } else {
    wc.Aec.Infof("Returning nil content - err no such entity")
    err = datastore.ErrNoSuchEntity
  }
  return nil, err

}

type Content struct {
  Key *datastore.Key `datastore:"-"`
  PageKey *datastore.Key
  Theme string
  Title string
  Draft bool
  // version is achieved with CreatedAt, or maybe not?
  Markdown string `datastore:",noindex"`
  CreatedAt time.Time
  Version int

  // non datastore fields
  Selected bool `datastore:"-"`
}

func (c *Content) Validate() map[string]string {
  ret := map[string]string{}

  if c.Theme == "" { ret["theme"] = "Please select a theme to use" }
  if c.Title == "" { ret["title"] = "Please enter a title for your page" }
  if c.Markdown == "" { ret["markdown"] = "Please enter some content" }
  if c.PageKey == nil { ret["site_id"] = "Please select a page" }
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
  draft := template.Must(template.ParseFiles(c.Theme))
  var output bytes.Buffer
  url := "http://"
  if wc.Ctx.HttpRequest().TLS != nil { url = "https://" }
  url = url + appengine.DefaultVersionHostname(wc.Aec) + "/" + path.Dir(c.Theme) + "/"
  data := struct {
    Content template.HTML
    Title string
    ThemePath string
  }{
    template.HTML(blackfriday.MarkdownBasic([]byte(c.Markdown))),
    c.Title,
    url,
  }
  draft.Execute(&output, data )
  return output
}
