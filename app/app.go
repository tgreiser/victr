package app

import (
  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"
  "github.com/stretchr/goweb/handlers"

  "github.com/tgreiser/victr/controllers"
)

func Init() {
  handler := handlers.NewHttpHandler(goweb.CodecService)

  //handler.Map("GET", "/", func (c context.Context) error {
  //  return goweb.Respond.With(c, 200, []byte("Hey planet, what's up"))
  //})

  content := new(controllers.ContentController)
  handler.MapController(content)
  handler.Map("GET", "/content/new", func (c context.Context) error {
    return content.New(c)
  })
  handler.Map("POST", "/content/publish", func (c context.Context) error {
    return content.Publish(c)
  })

  // failover handler
  /*
  handler.Map(func(c context.Context) error {
    return NotFound(c)
  })

  */
  http.Handle("/content", handler)
}

func NotFound(c context.Context) error {
  return goweb.Respond.With(c, 404, []byte("File not found"))
}

func SystemError(c context.Context) error {
  return goweb.Respond.With(c, 500, []byte("Server error, please try again in a moment."))
}
