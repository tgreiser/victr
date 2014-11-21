package models

import (
  "io/ioutil"

  mycontext "github.com/tgreiser/victr/context"
)

var (
  MAX_THEMES = 100
)

func FetchThemes(wc mycontext.Context, sel string) ([]*Theme, error) {

  // list everything in app/themes
  dir, err := ioutil.ReadDir("themes")
  themes := make([]*Theme, len(dir))
  if err != nil {
    wc.Aec.Errorf("unable to read themes dir: %v", err)
    return nil, err
  }

  for i, f := range dir {
    wc.Aec.Infof("Found %v %v", i, f)
    t := &Theme { Name: f.Name() }
    if t.Name == sel {
      t.Selected = true
    }
    wc.Aec.Infof("Listing theme: %v", t.Name)

    themes[i] = t
  }

  return themes, nil
}

type Theme struct {
  Name string
  Selected bool
}
