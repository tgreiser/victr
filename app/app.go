package app

import (
  "net/http"
  "path"

  "github.com/nicksnyder/go-i18n/i18n"
  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"
  "github.com/stretchr/goweb/handlers"

  mycontext "github.com/tgreiser/victr/context"
  "github.com/tgreiser/victr/controllers"
)

func init() {
  translation_path := mycontext.AppPath(path.Join("languages", "en_US.json"))
  i18n.MustLoadTranslationFile(translation_path)
  handler := handlers.NewHttpHandler(goweb.CodecService)

  //handler.Map("GET", "/", func (c context.Context) error {
  //  return goweb.Respond.With(c, 200, []byte("Hey planet, what's up"))
  //})

  content := new(controllers.ContentController)
  handler.MapController(content)
  handler.Map("POST", "/content/publish", func (c context.Context) error {
    return content.Publish(c)
  })
  handler.Map([]string {"GET","POST"}, "/content/preview/{site}/{path}", func (c context.Context) error {
    return content.Preview(c)
  })

  // failover handler
  /*
  handler.Map(func(c context.Context) error {
    return NotFound(c)
  })

  */
  http.Handle("/content/", handler)

  conf := handlers.NewHttpHandler(goweb.CodecService)
  sites := new(controllers.SitesController)
  conf.MapController(sites)
  conf.Map(func(c context.Context) error {
    wc := mycontext.NewContext(c)
    wc.Aec.Infof("Not found: %v %v", c.MethodString(), c.Path())
    return NotFound(c)
  })
  http.Handle("/sites/", conf)

  fileh := handlers.NewHttpHandler(goweb.CodecService)
  files := new(controllers.FilesController)
  fileh.MapController(files)
  fileh.Map(func(c context.Context) error {
    wc := mycontext.NewContext(c)
    wc.Aec.Infof("Not found: %v %v", c.MethodString(), c.Path())
    return NotFound(c)
  })
  http.Handle("/files/", fileh)
}

func NotFound(c context.Context) error {
  return goweb.Respond.With(c, 404, []byte("File not found"))
}

func SystemError(c context.Context) error {
  return goweb.Respond.With(c, 500, []byte("Server error, please try again in a moment."))
}
