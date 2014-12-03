package controllers

import (
  "fmt"
  "strconv"

  "appengine/datastore"
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
  return ctrl.renderSites(wc, wc.Ctx.FormValue("msg"), map[string]string{}, nil)
}

/*
This actually does a ReadMany action combined with Edit
*/
func (ctrl *SitesController) Read(key string, c context.Context) error {
  wc := mycontext.NewContext(c)
  wc.Aec.Infof("Rendering Config:Read %v", key)
  var site models.Site
  msg := ""
  k, err := datastore.DecodeKey(key)
  if err == nil { err = models.FindSite(wc, k, &site) }
  if err != nil {
    msg := "Unable to load site"
    wc.Aec.Errorf("%v: %v", msg, err)
  }
  return ctrl.renderSites(wc, msg, map[string]string{}, &site)
}

func (ctrl *SitesController) fetchSites(wc mycontext.Context) []*models.Site {
  limit, ce := strconv.Atoi(wc.Ctx.FormValue("limit"))
  if ce != nil { limit = 50 }
  offset, ce := strconv.Atoi(wc.Ctx.FormValue("offset"))
  if ce != nil { offset = 0 }
  sites, err := models.FetchSites(wc, limit, offset)
  if err != nil {
    wc.Aec.Errorf("error fetching sites, using empty list: %v", err)
    sites = make([]*models.Site, 0, 0)
  }
  return sites
}

func (ctrl *SitesController) fetchThemes(wc mycontext.Context, bucket, sel string) ([]*models.Theme, error) {
  themes, err := models.FetchThemes(wc, bucket, sel)
  if err != nil || len(themes) == 0 {
    wc.Aec.Errorf("error fetching themes, panic: %v", err)
    return nil, ctrl.error(wc, "err_no_themes", err)
  }
  wc.Aec.Infof("Returned %v themes", len(themes))
  return themes, nil
}

func (ctrl *SitesController) renderSites(wc mycontext.Context, message string, errs map[string]string, edit *models.Site) error {
  sites := ctrl.fetchSites(wc)
  sel := ""
  if (edit != nil) { sel = edit.Theme }
  themes, err := ctrl.fetchThemes(wc, edit.Bucket, sel)
  if err != nil || len(themes) == 0 {
    errs["theme"] = "No layout files were found at: site/themes/" + edit.Bucket + "/*.html"
  }
  wc.Aec.Infof("found %v sites", len(sites))
  wc.Aec.Infof("edit set? %v", edit != nil)
  wc.Aec.Infof("Errs: %v %v", len(errs), errs)
  if edit != nil {
    wc.Aec.Infof("edit: %v", edit)
  }

  data := struct {
    Message string
    Errors map[string]string
    Edit *models.Site
    Sites []*models.Site
    Themes []*models.Theme
  } {
    message,
    errs,
    edit,
    sites,
    themes,
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
    Theme: wc.Ctx.FormValue("theme"),
  }

  // validate
  wc.Aec.Infof("Validating...")
  if errs := site.Validate(); len(errs) > 0 {
    wc.Aec.Warningf("Failed to validate: %v %v", len(errs), errs)
    msg := "Failed to save"
    return ctrl.renderSites(wc, msg, errs, site)
  }

  // save
  wc.Aec.Infof("Saving...")
  if err := site.Save(wc, models.NewSiteKey(wc)); err != nil {
    msg := "Failed to save"
    wc.Aec.Errorf("msg %v", err)
    return ctrl.renderSites(wc, msg, map[string]string { }, site)
  }

  return goweb.Respond.WithRedirect(wc.Ctx, fmt.Sprintf("/sites/%s", site.Key.Encode()))
}

func (ctrl *SitesController) Delete(key string, c context.Context) error {
  wc := mycontext.NewContext(c)
  msg := ""
  var site models.Site
  k, err := datastore.DecodeKey(key)
  if err == nil { err = models.FindSite(wc, k, &site) }
  if err != nil {
    msg = "Unable to find site"
    wc.Aec.Errorf("%v: %v", msg, err)
  } else {
    if err = site.Delete(wc); err != nil {
      msg = "Unable to delete site"
      wc.Aec.Errorf("%v: %v", msg, err)
    }
  }
  return ctrl.renderSites(wc, msg, map[string]string{}, &site)
}
