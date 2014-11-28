package models

import (
  "strings"

  "appengine/datastore"
)

func NiceKey(key *datastore.Key) string {
  a := strings.Split(key.String(), ",")
  c := len(a)
  if c == 0 { return a[0] }
  return a[c-1]
}
