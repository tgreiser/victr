package models

import (
  "appengine"
  "appengine/datastore"
  "time"

  mycontext "github.com/tgreiser/victr/context"
)

func NewSiteKey(wc mycontext.Context) *datastore.Key {
  return datastore.NewIncompleteKey(wc.Aec, "Site", nil)
}

func FetchSites(wc mycontext.Context, limit, offset int) ([]*Site, error) {
  q := datastore.NewQuery("Site").Order("-CreatedAt").Limit(limit).Offset(offset)
  sites := make([]*Site, 0, limit)
  keys, err := q.GetAll(wc.Aec, &sites)
  if _, ok := err.(*datastore.ErrFieldMismatch); ok {
    wc.Aec.Infof("datastore missing field, ignoring: %v", err)
    err = nil
  } else if err != nil {
    wc.Aec.Errorf("got error instead of sites list: %v", err)
    return nil, err
  }

  for i, k := range keys {
    sites[i].Key = k
  }
  return sites, err
}

func FindSiteFromEnc(wc mycontext.Context, enc string, s *Site) error {
  k, err := datastore.DecodeKey(enc)
  if err != nil {
    wc.Aec.Errorf("failed to decode key: %v %v", enc, err)
    return err
  }
  return FindSite(wc, k, s)
}

func FindSite(wc mycontext.Context, k *datastore.Key, s *Site) error {
  if err := datastore.Get(wc.Aec, k, s); err != nil {
    if err != datastore.ErrNoSuchEntity {
      wc.Aec.Errorf("datastore error with FindSite: %v", err)
    }
    return err
  }
  s.Key = k
  return nil
}

type Site struct {
  Key *datastore.Key `datastore:"-"`
  Name string
  URL string
  Bucket string
  Theme string
  CreatedAt time.Time
  UpdatedAt time.Time
  ImagePath string `datastore:",noindex"`
}

func (s *Site) Validate() map[string]string {
  ret := map[string]string{}

  if s.Name == "" {
    ret["name"] = "Please enter the site name"
  }
  if s.URL == "" {
    ret["url"] = "Please enter the site URL"
  }
  if s.Bucket == "" {
    ret["bucket"] = "Please enter the site bucket"
  }
  if s.Theme == "" {  ret["theme"] = "Please pick a default theme" }
  if s.ImagePath == "" { ret["image_path"] = "Please enter an image upload path" }

  return ret
}

func (s *Site) Save(wc mycontext.Context, key *datastore.Key) error {
  err := datastore.RunInTransaction(wc.Aec, func(aec appengine.Context) error {
    if s.Key == nil { s.CreatedAt = time.Now() }
    s.UpdatedAt = time.Now()
    key, e := datastore.Put(aec, key, s)
    if e != nil {
      return e
    }
    s.Key = key

    return nil
  }, nil)
  if err != nil {
    wc.Aec.Errorf("datastore write failed: %v", err)
  }
  return err
}

func (s *Site) Delete(wc mycontext.Context) error {
  err := datastore.RunInTransaction(wc.Aec, func(aec appengine.Context) error {
    return datastore.Delete(aec, s.Key)
  }, nil)
  if err != nil {
    wc.Aec.Errorf("datastore delete failed: %v", err)
  }
  return err
}
