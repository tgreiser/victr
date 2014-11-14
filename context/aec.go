package context

import(
  "appengine"
  "net/http"
)

func Aec(req *http.Request) appengine.Context {
  aec := appengine.NewContext(req)

  return aec
}
