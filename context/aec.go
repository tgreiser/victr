package context

import(
  "appengine"
  "net/http"
)

func Aec(req *http.Request) appengine.Context {
  aec := appengine.NewContext(req)
  ns := req.Header.Get("X-Namespace")
  if ns == "" {
    ns = "prod"
  }

  var err error
  aec.Infof("Setting namespace: %v", ns)
  aec, err = appengine.Namespace(aec, ns)
  if err != nil {
    aec = appengine.NewContext(req)
    aec.Errorf("Error loading namespace!")
  }

  return aec
}
