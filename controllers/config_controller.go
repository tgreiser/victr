package controllers

import (
  "github.com/stretchr/goweb/context"

  mycontext "github.com/tgreiser/victr/context"
)

type ConfigController struct {
  BaseController
}

func (ctrl *ConfigController) ReadMany(c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Rendering Config:ReadMany")
  return ctrl.render(wc, "config", "")
}
