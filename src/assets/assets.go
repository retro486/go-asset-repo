package assets

import (
  "auth"
  "net/http"
  "html/template"
  "fmt"
)

type Asset struct {
  URL string
  Name string
  IsImage bool // for determining if a thumbnail preview should be shown
}

func LoadStoredAssets() []Asset {
  // TODO
  assets := []Asset{}
  assets = append(assets, Asset{URL:"http://www.google.com", Name:"Google", IsImage:false})
  assets = append(assets, Asset{URL:"http://www.yahoo.com", Name:"Yahoo", IsImage:false})
  assets = append(assets, Asset{URL:"http://minionslovebananas.com/images/check-in-minion.jpg", Name:"Minion", IsImage:true})

  return assets
}

func ShowIndex(w http.ResponseWriter, r *http.Request) {
  error := false
  auth.CheckAuthCookie(w, r)
  tmpl, _ := template.ParseFiles("views/assets/index.html")
  assets := LoadStoredAssets()
  err := tmpl.Execute(w, map[string]interface{} { "Assets": assets })
  if err != nil {
    fmt.Printf("Unable to write template to response.\n")
    error = true
  }

  if error {
    http.ServeFile(w, r, "views/error.html")
  }
}
