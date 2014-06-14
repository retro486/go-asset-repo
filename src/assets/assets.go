package assets

import (
  "github.com/gorilla/mux"
  "auth"
  "net/http"
  "html/template"
  "fmt"
  "log"
  "database/sql"
  "strconv"
  _ "github.com/mattn/go-sqlite3"
)

type Asset struct {
  Id int64
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

  _, err = db.Exec(dbCreateSQL)
  if err == nil {
    log.Printf("Created database")
  }

  return db
}

// Stores the given asset in the db. Returns nil if fail, the given Asset with Id if success.
func CreateAsset(asset Asset) (Asset, error) {
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

  result, err := stmt.Exec(asset.Name, asset.URL, asset.IsImage)
  if err != nil {
    return Asset{}, err
  }

  tx.Commit()

  asset.Id, _ = result.LastInsertId()
  return asset, nil
}

func DestroyAsset(id int64) error {
  dbConn := GetDBConnection()
  defer dbConn.Close()

  tx, err := dbConn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("delete from assets where id = ?")
  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec(id)
  if err != nil {
    return err
  }

  tx.Commit()

  return nil
}

func LoadStoredAssets() []Asset {
  dbConn := GetDBConnection()
  defer dbConn.Close()
  assets := []Asset{}

  // _, err := CreateAsset(Asset{URL:"http://www.google.com", Name:"Google", IsImage:false})
  // if err != nil {
  //   fmt.Printf("%s\n", fmt.Errorf("%v", err))
  // }
  // _, err = CreateAsset(Asset{URL:"http://www.yahoo.com", Name:"Yahoo", IsImage:false})
  // if err != nil {
  //   fmt.Printf("%s\n", fmt.Errorf("%v", err))
  // }
  // _, err = CreateAsset(Asset{URL:"http://minionslovebananas.com/images/check-in-minion.jpg", Name:"Minion", IsImage:true})
  // if err != nil {
  //   fmt.Printf("%s\n", fmt.Errorf("%v", err))
  // }

  rows, err := dbConn.Query("select id, name, url, isimage from assets")
  if err != nil {
    return nil
  }
  defer rows.Close()

  for rows.Next() {
    var id int64
    var name string
    var url string
    var isImage bool
    rows.Scan(&id, &name, &url, &isImage)
    assets = append(assets, Asset{Id: id, URL: url, Name: name, IsImage: isImage})
  }
  rows.Close()

  return assets
}

func ControllerDestroyAsset(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, _ := strconv.ParseInt(vars["id"], 10, 32)

  _ = DestroyAsset(id)
  http.Redirect(w, r, "/assets", 302)
}

func ControllerShowIndex(w http.ResponseWriter, r *http.Request) {
  error := false
  auth.CheckAuthCookie(w, r)
  assets := LoadStoredAssets()

  tmpl, _ := template.ParseFiles("views/assets/index.html")

  err := tmpl.Execute(w, map[string]interface{} { "Assets": assets })
  if err != nil {
    fmt.Printf("Unable to write template to response.\n")
    error = true
  }

  if error {
    http.ServeFile(w, r, "views/error.html")
  }
}
