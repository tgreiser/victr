package controllers

import (
  "fmt"
  "strconv"

  "github.com/stretchr/goweb"
  "github.com/stretchr/goweb/context"

  mycontext "github.com/tgreiser/victr/context"
  "github.com/tgreiser/victr/models"
)

type SitesController struct {
  BaseController
}

func (ctrl *SitesController) ReadMany(c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Rendering Config:ReadMany")
  return ctrl.renderSites(wc, "", map[string]string{}, nil)
}

func (ctrl *SitesController) renderSites(wc mycontext.Context, message string, errs map[string]string, edit *models.Site) error {
  limit, ce := strconv.Atoi(wc.Ctx.FormValue("limit"))
  if ce != nil { limit = 50 }
  offset, ce := strconv.Atoi(wc.Ctx.FormValue("offset"))
  if ce != nil { offset = 0 }
  sites, err := models.FetchSites(wc, limit, offset)
  if err != nil {
    wc.Aec.Errorf("error fetching sites, using empty list: %v", err)
    sites = make([]*models.Site, 0, 0)
  }
  wc.Aec.Infof("found %v sites", len(sites))

  data := struct {
    Message string
    Errors map[string]string
    Edit *models.Site
    Sites []*models.Site
  } {
    message,
    errs,
    edit,
    sites,
  }
  return ctrl.render(wc, "sites", data)
}

func (ctrl *SitesController) Create(c context.Context) error {
  wc := mycontext.NewContext(c)

  wc.Aec.Infof("Running sites:create")

  // read form vars
  site := &models.Site{
    Name: wc.Ctx.FormValue("name"),
    URL: wc.Ctx.FormValue("url"),
    Bucket: wc.Ctx.FormValue("bucket"),
  }

  // validate
  wc.Aec.Infof("Validating...")
  if errs := site.Validate(); len(errs) > 0 {
    msg := "Failed to save"
    ctrl.renderSites(wc, msg, errs, site)
  }

  // save
  wc.Aec.Infof("Saving...")
  if err := site.Save(wc, models.NewSiteKey(wc, site.URL)); err != nil {
    msg := "Failed to save"
    wc.Aec.Errorf("msg %v", err)
    ctrl.renderSites(wc, msg, map[string]string { }, site)
  }

  return goweb.Respond.WithRedirect(wc.Ctx, fmt.Sprintf("/sites/%s", site.Key.Encode()))
}
