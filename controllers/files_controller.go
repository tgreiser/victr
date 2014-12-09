package controllers

import (
  "bytes"
  "io/ioutil"

  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"

  mycontext "github.com/tgreiser/victr/context"
  "github.com/tgreiser/victr/storage"
)

type FilesController struct {
  BaseController
}

func (ctrl *FilesController) ReadMany(c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Rendering Config:ReadMany")
  return ctrl.renderReadMany(wc, "", map[string]string{})
}

func (ctrl *FilesController) renderReadMany(wc mycontext.Context, msg string, errs map[string]string ) error {
  bucket := wc.Ctx.FormValue("bucket")
  path := wc.Ctx.FormValue("path")
  data := struct {
    Bucket string
    Path string
  } {
    bucket,
    path,
  }

  return ctrl.render(wc, "files", data)
}

func (ctrl *FilesController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)
  r := wc.Ctx.HttpRequest()

  err := r.ParseMultipartForm(2000000)
  if err != nil { return ctrl.err(wc, err) }
  m := r.MultipartForm

  bucket := m.Value["bucket"][0]
  path := m.Value["path"][0]
  wc.Aec.Infof("Files Create %v %v", bucket, path)
  wc.Aec.Infof("%v", wc.Ctx.FormParams())
  file, handler, err := wc.Ctx.HttpRequest().FormFile("image")
  if err != nil {
    return ctrl.err(wc, err)
  }
  wc.Aec.Infof("Reading...")
  data, err := ioutil.ReadAll(file)
  if err != nil {
    return ctrl.err(wc, err)
  }
  // write data to cloud storage
  obj, err := storage.NewObject(wc, bucket, path + "/" + handler.Filename)
  if err == nil {
    err = obj.Store(wc, bytes.NewReader(data))
  }
  if err != nil {
    return ctrl.err(wc, err)
  }

  return ctrl.returnOK(wc)
}

func (ctrl *FilesController) err(wc mycontext.Context, err error) error {
  wc.Aec.Errorf("Files: %v", err)
  data := struct {
    Success bool
  } {
    false,
  }

  return goweb.API.RespondWithData(wc.Ctx, data)
}

func (ctrl *FilesController) returnOK(wc mycontext.Context) error {
  data := struct {
    Success bool
  } {
    true,
  }

  return goweb.API.RespondWithData(wc.Ctx, data)
}
