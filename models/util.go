package models

import (
  "strconv"
  "strings"

  "appengine/datastore"
  mycontext "github.com/tgreiser/victr/context"
)

func NiceKey(key *datastore.Key) string {
  a := strings.Split(key.String(), ",")
  c := len(a)
  if c == 0 { return a[0] }
  return a[c-1]
}

func DsKey(wc mycontext.Context, entity, nice string) *datastore.Key {
  key, err := strconv.ParseInt(nice, 10, 64)
  if err != nil {
    wc.Aec.Warningf("Unable to parse %v key %v: %v", entity, nice, err)
    return nil
  }
  return datastore.NewKey(wc.Aec, entity, "", key, nil)
}
