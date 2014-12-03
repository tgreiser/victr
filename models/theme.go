package models

import (
  "path"
  "path/filepath"

  mycontext "github.com/tgreiser/victr/context"
)

var (
  MAX_THEMES = 100
)

func FetchThemes(wc mycontext.Context, bucket, sel string) ([]*Theme, error) {

  // list everything in app/themes
  dir, err := filepath.Glob(path.Join("themes", bucket, "*.html"))
  if err != nil {
    wc.Aec.Errorf("unable to read themes dir: %v", err)
    return nil, err
  }
  themes := make([]*Theme, len(dir))

  for i, f := range dir {
    wc.Aec.Infof("Found %v %v %v", i, f)
    t := &Theme { Path: f, Name: path.Base(f) }
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
  Path string
  Selected bool
}
