package models

import (
  "appengine"
  "appengine/datastore"
  "time"

  mycontext "github.com/tgreiser/victr/context"
)

func NewPageKey(wc mycontext.Context) *datastore.Key {
  return datastore.NewIncompleteKey(wc.Aec, "Page", nil)
}

func NewPage(wc mycontext.Context) *Page {
  page := &Page {
    Path: wc.Ctx.FormValue("path"),
    CurrentVersion: 1,
  }
  sitekey, err := datastore.DecodeKey(wc.Ctx.FormValue("site_id"))
  if err != nil {
    wc.Aec.Warningf("Failed to decode site key: %v %v", wc.Ctx.FormValue("site_id"), err)
    return page
  }
  page.SiteKey = sitekey
  return page
}

func FindPage(wc mycontext.Context, k *datastore.Key, p *Page) error {
  if err := datastore.Get(wc.Aec, k, p); err != nil {
    if err != datastore.ErrNoSuchEntity {
      wc.Aec.Errorf("datastore error with FindPage: %v", err)
    }
    return err
  }
  p.Key = k
  wc.Aec.Infof("Page found: %v", p)
  return nil
}

func FindPageByPath(wc mycontext.Context, site_key *datastore.Key, path string, p *Page) error {
  q := datastore.NewQuery("Page").Filter("Path=", path).Filter("SiteKey=", site_key).Limit(1)
  pages := make([]*Page, 1, 1)
  keys, err := q.GetAll(wc.Aec, &pages)
  if _, ok := err.(*datastore.ErrFieldMismatch); ok {
    wc.Aec.Infof("datastore missing field, ignoring: %v", err)
    err = nil
  } else if err != nil {
    wc.Aec.Errorf("got error instead of page: %v", err)
    return err
  } else if len(keys) > 0 {
    p = pages[0]
    wc.Aec.Infof("Got keys: %v", keys)
    wc.Aec.Infof("Got Pages: %v", pages)
    p.Key = keys[0]
  } else {
    wc.Aec.Infof("Returning nil page")
    p = nil
    err = datastore.ErrNoSuchEntity
  }
  return err
}

type Page struct {
  Key *datastore.Key `datastore:"-"`
  SiteKey *datastore.Key
  Path string
  CurrentVersion int
  CurrentVersionKey *datastore.Key
  CreatedAt time.Time
  UpdatedAt time.Time
  Published bool
}

func (p *Page) Validate(wc mycontext.Context) map[string]string {
  ret := map[string]string{}

  if p.Path == "" { ret["path"] = "Please enter the relative path where your file will be published" }
  if p.SiteKey == nil { ret["site"] = "Please select a site" }

  if p.CurrentVersion == 1 {
    // first save, verify the page isn't a name conflict
    var page Page
    err := FindPageByPath(wc, p.SiteKey, p.Path, &page)
    if err != nil && err != datastore.ErrNoSuchEntity {
      wc.Aec.Errorf("path lookup err: %v", err)
      ret["path"] = "An error occurred when validating your path. Is it correct? (<something>.html)"
    } else if err == nil {
      wc.Aec.Infof("Found page: %v", page)
      ret["path"] = "You entered a path for an existing page, copy your content, then " +
        "<a href=\"/content/" + page.Key.StringID() + "\">click here.</a>"
    }
  }

  return ret
}

func (p *Page) Save(wc mycontext.Context, key *datastore.Key) error {
  err := datastore.RunInTransaction(wc.Aec, func(aec appengine.Context) error {
    if p.Key == nil { p.CreatedAt = time.Now() }
    p.UpdatedAt = time.Now()
    key, e := datastore.Put(aec, key, p)
    if e != nil {
      return e
    }
    p.Key = key
    return nil
  }, nil)
  if err != nil {
    wc.Aec.Errorf("datastore write failed: %v", err)
  }
  return err
}

func (p *Page) LiveUrl(wc mycontext.Context) string {
  var site Site
  if e := FindSite(wc, p.SiteKey, &site); e != nil {
    wc.Aec.Errorf("error building URL, no site: %v", e)
    return "#"
  }
  return site.URL + "/" + p.Path
}
