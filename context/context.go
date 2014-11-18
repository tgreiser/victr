package context

import (
  "appengine"
  "fmt"
  "path"
  "runtime"

  "github.com/nicksnyder/go-i18n/i18n"
  "github.com/stretchr/goweb/context"
)

func NewContext(c context.Context) Context {
  ctx := Context {
    Aec: Aec(c.HttpRequest()),
    Ctx: c,
  }
  var err error
  ctx.T, err = LoadTranslator(ctx.Aec, ctx.GetLocale())
  if err != nil {
    ctx.Aec.Errorf("Error loading translator: %v %v", ctx.GetLocale(), err)
  }
  return ctx
}

type Context struct {
  Aec appengine.Context
  Ctx context.Context
  T i18n.TranslateFunc
}

func (c *Context) GetLocale() string {
  // TODO : check the headers or cookie or whatever
  return "en_US"
}

func AppPath(filename_from_app string) string {
  _, filename, _, _ := runtime.Caller(1);
  return path.Join(path.Dir(filename), "..", "app", filename_from_app)
}

func LoadTranslator(aec appengine.Context, locale string) (i18n.TranslateFunc, error) {
  if locale == "" {
    locale = "en_US"
  }
  T, err := i18n.Tfunc(locale)
  if err != nil {
    aec.Errorf("Could not load a valid language file for %v! Had to send a static english error message %v", locale, err)
    if locale != "en_US" {
      T, err = LoadTranslator(aec, "en_US")
    }
    if err != nil {
      return nil, fmt.Errorf("Failed to load requested language file, or English backup:%v", err)
    }
  }

  return T, err
}
