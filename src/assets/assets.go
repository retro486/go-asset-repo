package assets

import (
  "github.com/gorilla/mux"
  "github.com/gorilla/schema"
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
  Id *int64
  URL string `schema:"url"`
  Name string `schema:"name"`
  IsImage bool `schema:"isimage"` // for determining if a thumbnail preview should be shown
}

func GetDBConnection() *sql.DB {
  dbCreateSQL := "create table assets (id integer not null primary key autoincrement, name text, url text, isimage boolean);"
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

  id, _ := result.LastInsertId()
  asset.Id = &id
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
    assets = append(assets, Asset{Id: &id, URL: url, Name: name, IsImage: isImage})
  }
  rows.Close()

  return assets
}

func FindAsset(id int64) Asset {
  dbConn := GetDBConnection()
  defer dbConn.Close()

  stmt, err := dbConn.Prepare("select id, name, url, isimage from assets where id = ?")
  if err != nil {
    return Asset{}
  }
  defer stmt.Close()

  var name string
  var url string
  var isImage bool
  err = stmt.QueryRow(id).Scan(&id, &name, &url, &isImage)
  if err != nil {
    return Asset{}
  }

  return Asset{Id: &id, URL: url, Name: name, IsImage: isImage}
}

func UpdateAsset(asset *Asset) error {
  dbConn := GetDBConnection()
  defer dbConn.Close()

  tx, err := dbConn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("update assets set name = ?, url = ?, isimage = ? where id = ?")
  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec(asset.Name, asset.URL, asset.IsImage, &asset.Id)
  if err != nil {
    return err
  }

  tx.Commit()

  return nil
}

func ControllerEditAsset(w http.ResponseWriter, r *http.Request) {
  error := false
  vars := mux.Vars(r)
  id, _ := strconv.ParseInt(vars["id"], 10, 32)
  asset := FindAsset(id)
  if asset.Id == nil {
    // ERROR asset not found or DB error
    http.Redirect(w, r, "/assets", 302)
  }

  tmpl, _ := template.ParseFiles("views/assets/edit.html")

  err := tmpl.Execute(w, asset)
  if err != nil {
    fmt.Printf("Unable to write edit template to response.\n")
    error = true
  }

  if error {
    http.ServeFile(w, r, "views/error.html")
  }
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
    fmt.Printf("Unable to write index template to response.\n")
    error = true
  }

  if error {
    http.ServeFile(w, r, "views/error.html")
  }
}

func ControllerUpdateAsset(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
  id, _ := strconv.ParseInt(vars["id"], 10, 32)

  if r.Method == "POST" {
    err := r.ParseForm()
    if err != nil {
      // ERROR: Unable to read form data.
      fmt.Printf("ERROR: Bad form data\n")
      http.Redirect(w, r, "/", 302)
    } else {
      asset := new(Asset)
      decoder := schema.NewDecoder()

      err := decoder.Decode(asset, r.PostForm)
      if err == nil {
        asset.Id = &id
        _ = UpdateAsset(asset)
      }

      http.Redirect(w, r, "/assets", 302)
    }
  } else {
    http.Redirect(w, r, "/assets", 302)
  }
}
