package models

import (
  "appengine"
  "appengine/datastore"

  mycontext "github.com/tgreiser/victr/context"
)

func NewSiteKey(wc mycontext.Context, url string) *datastore.Key {
  return datastore.NewKey(wc.Aec, "Site", url, 0, nil)
}

func FetchSites(wc mycontext.Context, limit, offset int) ([]*Site, error) {
  q := datastore.NewQuery("Site").Order("Name").Limit(limit).Offset(offset)
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

type Site struct {
  Key *datastore.Key
  Name string
  URL string
  Bucket string
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

  return ret
}

func (s *Site) Save(wc mycontext.Context, key *datastore.Key) error {
  err := datastore.RunInTransaction(wc.Aec, func(aec appengine.Context) error {
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
