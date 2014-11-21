package models

import (
  "io/ioutil"

  mycontext "github.com/tgreiser/victr/context"
)

var (
  MAX_THEMES = 100
)

func FetchThemes(wc mycontext.Context) ([]*Theme, error) {
  themes := make([]*Theme, 0, MAX_THEMES)

  // list everything in app/themes
  dir, err := ioutil.ReadDir("themes")
  if err != nil {
    wc.Aec.Errorf("unable to read themes dir: %v", err)
    return nil, err
  }

  for f := range dir {
    wc.Aec.Infof("Found %v", f)
  }

  return themes, nil
}

type Theme struct {
  Name string
}
