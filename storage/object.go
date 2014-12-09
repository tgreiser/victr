package storage

import (
  "io"
  "net/http"

  gstorage "code.google.com/p/google-api-go-client/storage/v1"
  "github.com/golang/oauth2"
  "github.com/golang/oauth2/google"

  mycontext "github.com/tgreiser/victr/context"
)

func NewObject(wc mycontext.Context, bucket, path string) (*Object, error) {
  // get oauth client
  wc.Aec.Infof("Getting oauth client")
  f, err := oauth2.New(
    google.AppEngineContext(wc.Aec),
    oauth2.Scope(
      "https://www.googleapis.com/auth/devstorage.read_write",
    ),
  )
  if err != nil {
    wc.Aec.Errorf("cloud storage auth failed: %v", err)
    // TODO return to /content/new with pagedata pre-filled
    return nil, err
  }
  client := http.Client{Transport: f.NewTransport()}

  // do the cloud storage put operation
  wc.Aec.Infof("Cloud storage put...")
  storeSvc, err := gstorage.New(&client)
  if err != nil {
    wc.Aec.Errorf("failed to get storage client: %v", err)
    return nil, err
  }
  obj := gstorage.NewObjectsService(storeSvc)
  object := &gstorage.Object {
    Bucket: bucket,
    ContentType: "text/html",
    Name: path,
  }

  return &Object{
    Service: obj,
    Inner: object,
  }, nil
}

type Object struct {
  Service *gstorage.ObjectsService
  Inner *gstorage.Object
}

func (ob *Object) List(wc mycontext.Context) ([]string, error) {
  objs, err := ob.Service.List(ob.Inner.Bucket).Do()
  var found = []string{}
  for _, e := range objs.Items {
    wc.Aec.Infof("Matching %v and %v", e.Name, ob.Inner.Name)
    if len(e.Name) > len(ob.Inner.Name) && e.Name[0:len(ob.Inner.Name)] == ob.Inner.Name {
      found = append(found, e.Name)
    } else {
      wc.Aec.Infof("%v no match for %v", e.Name, ob.Inner.Name)
    }
  }

  return found, err
}

func (ob *Object) Store(wc mycontext.Context, r io.Reader) error {
  _,  err := ob.Service.Insert(ob.Inner.Bucket, ob.Inner).Media(r).Do()
  if err != nil {
    wc.Aec.Errorf("Failed to store object: %v %v", ob.Inner, err)
  }
  return err
}
