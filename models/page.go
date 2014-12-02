package models

import (
  "appengine"
  "appengine/datastore"
  "strconv"
  "time"

  mycontext "github.com/tgreiser/victr/context"
)

func NewPageKey(wc mycontext.Context) *datastore.Key {
  return datastore.NewIncompleteKey(wc.Aec, "Page", nil)
}

func NewPage(wc mycontext.Context) *Page {
  wc.Aec.Infof("New page formvals: %v", wc.Ctx.FormParams())
  path := wc.Ctx.FormValue("path")
  page := &Page {
    Path: path,
    CurrentVersion: 1,
  }
  sitekey, err := datastore.DecodeKey(wc.Ctx.FormValue("site_id"))
  if err != nil {
    wc.Aec.Warningf("Failed to decode site key: %v %v", wc.Ctx.FormValue("site_id"), err)
    return page
  }
  wc.Aec.Infof("Site key: %v", sitekey)

  ver := wc.Ctx.FormValue("last_version")
  if ver == "" {
    page.SiteKey = sitekey
    page.Init(wc)
    page.CurrentVersion = 1
    page.Key = NewPageKey(wc)
  } else {
    page, err = FetchPageByPath(wc, sitekey, path)
    if err != nil {
      wc.Aec.Errorf("Failed to load path: %v", err)
      return nil
    }
    cv, err := strconv.Atoi(ver)
    if err == nil {
      page.CurrentVersion = cv+1
    }
  }
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
  p.Init(wc)
  wc.Aec.Infof("Page found: %v", p)
  return nil
}

func FetchPageByPath(wc mycontext.Context, site_key *datastore.Key, path string) (*Page, error) {
  wc.Aec.Infof("Fetch: path=%v site_key=%v", path, site_key)
  q := datastore.NewQuery("Page").Filter("Path=", path).Filter("SiteKey=", site_key).Limit(1)
  pages := make([]*Page, 0, 1)
  keys, err := q.GetAll(wc.Aec, &pages)
  if _, ok := err.(*datastore.ErrFieldMismatch); ok {
    wc.Aec.Infof("datastore missing field, ignoring: %v", err)
    err = nil
  } else if err != nil {
    wc.Aec.Errorf("got error instead of page: %v", err)
    return nil, err
  } else if len(keys) > 0 {
    pages[0].Key = keys[0]
    pages[0].Init(wc)
    wc.Aec.Infof("Loaded page: %v", pages[0])
    return pages[0], nil
  } else {
    wc.Aec.Infof("Returning nil page - err no such entity")
    err = datastore.ErrNoSuchEntity
  }
  return nil, err
}

func FetchPages(wc mycontext.Context, site_key *datastore.Key, limit, offset int) ([]*Page, error) {
  q := datastore.NewQuery("Page").
    Order("-UpdatedAt").
    Limit(limit).
    Offset(offset)
  pages := make([]*Page, 0, limit)
  keys, err := q.GetAll(wc.Aec, &pages)
  if _, ok := err.(*datastore.ErrFieldMismatch); ok {
    wc.Aec.Infof("datastore missing field, ignoring: %v", err)
    err = nil
  } else if err != nil {
    wc.Aec.Errorf("got error instead of content list: %v", err)
    return nil, err
  }

  for i, k := range keys {
    pages[i].Key = k
    pages[i].Init(wc)
    wc.Aec.Infof("Pages: %v", pages[i])
  }
  return pages, err
}


type Page struct {
  Key *datastore.Key `datastore:"-"`
  SiteKey *datastore.Key
  Path string
  CurrentVersion int
  CurrentVersionKey *datastore.Key
  MaxVersion int
  CreatedAt time.Time
  UpdatedAt time.Time
  Published bool
  Url string `datastore:"-"`
  NiceKey string `datastore:"-"`
}

func (p *Page) Init(wc mycontext.Context) {
  p.NiceKey = NiceKey(p.Key)
  p.Url = p.LiveUrl(wc)
}

func (p *Page) Validate(wc mycontext.Context) map[string]string {
  ret := map[string]string{}

  if p.Path == "" { ret["path"] = "Please enter the relative path where your file will be published" }
  if p.SiteKey == nil { ret["site_id"] = "Please select a site" }

  if p.CurrentVersion == 1 {
    // first save, verify the page isn't a name conflict
    page, err := FetchPageByPath(wc, p.SiteKey, p.Path)
    if err != nil && err != datastore.ErrNoSuchEntity {
      wc.Aec.Errorf("path lookup err: %v", err)
      ret["path"] = "An error occurred when validating your path. Is it correct? (<something>.html)"
    } else if err == nil {
      skey := NiceKey(page.Key)
      wc.Aec.Infof("Found page: %v %v", page.Key, skey)
      ret["path"] = "You entered a path for an existing page, copy your content, then " +
        "<a href=\"/content/" + skey + "\">click here.</a>"
    }
  }

  return ret
}

func (p *Page) Save(wc mycontext.Context, key *datastore.Key) error {
  err := datastore.RunInTransaction(wc.Aec, func(aec appengine.Context) error {
    if p.Key == nil { p.CreatedAt = time.Now() }
    p.UpdatedAt = time.Now()
    if p.CurrentVersion > p.MaxVersion { p.MaxVersion = p.CurrentVersion }
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
  wc.Aec.Infof("Page LiveURL: %v", p)
  var site Site
  if e := FindSite(wc, p.SiteKey, &site); e != nil {
    wc.Aec.Errorf("error building URL, no site: %v", e)
    return "#"
  }
  return site.URL + "/" + p.Path
}
