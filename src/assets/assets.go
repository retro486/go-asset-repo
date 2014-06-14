package assets

import (
  "auth"
  "net/http"
  "html/template"
  "fmt"
  "log"
  "errors"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

type Asset struct {
  URL string
  Name string
  IsImage bool // for determining if a thumbnail preview should be shown
}

var dbCreateSQL = "create table assets (id integer not null primary key autoincrement, name text, url text, isimage boolean);"
// var dbConn = *sql.DB{}

func GetDBConnection() *sql.DB {
  db, err := sql.Open("sqlite3", "assets.db")
  if err != nil {
    return nil
  }
  defer db.Close()

  // TODO if new database then run this
  _, err = db.Exec(dbCreateSQL)
  if err == nil {
    log.Printf("Created database")
  }

  return db
}

// Stores the given asset in the db. Returns nil if fail, the given Asset if success.
func SaveAsset(asset Asset) (Asset, error) {
  dbConn := GetDBConnection()

  tx, err := dbConn.Begin()
  if err != nil {
    return Asset{}, errors.New("Unable to start db transaction.")
  }

  stmt, err := tx.Prepare("insert into assets(name, url, isimage) values (?, ?, ?)")
  if err != nil {
    return Asset{}, errors.New("Unable to start db transaction.")
  }
  defer stmt.Close()

  _, err = stmt.Exec(asset.Name, asset.URL, asset.IsImage)
  if err != nil {
    return Asset{}, errors.New("Unable to start db transaction.")
  }

  tx.Commit()

  return asset, nil
}

func LoadStoredAssets() []Asset {
  // TODO
  dbConn := GetDBConnection()
  assets := []Asset{}

  rows, err := dbConn.Query("select * from assets")
  if err != nil {
    return nil
  }
  defer rows.Close()

  for rows.Next() {
    asset := Asset{}
    rows.Scan(&asset.Name, &asset.URL, &asset.IsImage)
    assets = append(assets, asset)
  }
  rows.Close()

  // assets = append(assets, Asset{URL:"http://www.google.com", Name:"Google", IsImage:false})
  // assets = append(assets, Asset{URL:"http://www.yahoo.com", Name:"Yahoo", IsImage:false})
  // assets = append(assets, Asset{URL:"http://minionslovebananas.com/images/check-in-minion.jpg", Name:"Minion", IsImage:true})

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
