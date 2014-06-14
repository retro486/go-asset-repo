package assets

import (
  "auth"
  "net/http"
  "html/template"
  "fmt"
  "log"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

type Asset struct {
  Id int
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
  defer dbConn.Close()

  tx, err := dbConn.Begin()
  if err != nil {
    return Asset{}, err
  }

  stmt, err := tx.Prepare("insert into assets(name, url, isimage) values (?, ?, ?)")
  if err != nil {
    return Asset{}, err
  }
  defer stmt.Close()

  _, err = stmt.Exec(asset.Name, asset.URL, asset.IsImage)
  if err != nil {
    return Asset{}, err
  }

  tx.Commit()

  return asset, nil
}

func LoadStoredAssets() []Asset {
  // TODO
  dbConn := GetDBConnection()
  defer dbConn.Close()
  assets := []Asset{}

  // _, err := SaveAsset(Asset{URL:"http://www.google.com", Name:"Google", IsImage:false})
  // if err != nil {
  //   fmt.Printf("%s\n", fmt.Errorf("%v", err))
  // }
  // _, err = SaveAsset(Asset{URL:"http://www.yahoo.com", Name:"Yahoo", IsImage:false})
  // if err != nil {
  //   fmt.Printf("%s\n", fmt.Errorf("%v", err))
  // }
  // _, err = SaveAsset(Asset{URL:"http://minionslovebananas.com/images/check-in-minion.jpg", Name:"Minion", IsImage:true})
  // if err != nil {
  //   fmt.Printf("%s\n", fmt.Errorf("%v", err))
  // }

  rows, err := dbConn.Query("select id, name, url, isimage from assets")
  if err != nil {
    return nil
  }
  defer rows.Close()

  for rows.Next() {
    var id int
    var name string
    var url string
    var isImage bool
    rows.Scan(&id, &name, &url, &isImage)
    assets = append(assets, Asset{Id: id, URL: url, Name: name, IsImage: isImage})
  }
  rows.Close()

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
