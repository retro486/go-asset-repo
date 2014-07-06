package assets

import (
  "github.com/gorilla/mux"
  "github.com/gorilla/schema"
  "auth"
  "net/http"
  "mime"
  "io/ioutil"
  "mime/multipart"
  "html/template"
  "fmt"
  "os"
  "strings"
  "log"
  "database/sql"
  "strconv"
  "net/url"
  _ "github.com/mattn/go-sqlite3"
)

type Asset struct {
  Id *int64 `schema:"-"`
  URL string `schema:"url"`
  Name string `schema:"name"`
  FileName string `schema:"filename"`
  IsImage bool `schema:"isimage"` // for determining if a thumbnail preview should be shown
}

var formDecoder = schema.NewDecoder()

func GetDBConnection() *sql.DB {
  dbCreateSQL := "create table assets (id integer not null primary key autoincrement, name text, url text, filename text, isimage boolean);"
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

  stmt, err := dbConn.Prepare("select id, name, url, filename, isimage from assets where id = ?")
  if err != nil {
    return Asset{}
  }
  defer stmt.Close()

  var name string
  var url string
  var fileName string
  var isImage bool
  err = stmt.QueryRow(id).Scan(&id, &name, &url, &fileName, &isImage)
  if err != nil {
    return Asset{}
  }

  return Asset{Id: &id, URL: url, Name: name, FileName: fileName, IsImage: isImage}
}

func CreateAsset(asset *Asset) error {
  dbConn := GetDBConnection()
  defer dbConn.Close()

  tx, err := dbConn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("insert into assets(name, url, filename, isimage) values (?, ?, ?, ?)")
  if err != nil {
    return err
  }
  defer stmt.Close()

  result, err := stmt.Exec(asset.Name, asset.URL, asset.FileName, asset.IsImage)
  if err != nil {
    return err
  }

  tx.Commit()

  id, _ := result.LastInsertId()
  asset.Id = &id
  return nil
}

func DestroyAsset(id int64) error {
  asset := FindAsset(id)

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
  } else {
    err = os.Remove(asset.FileName)
    if err != nil {
      fmt.Printf("%q:%q\n", err, asset)
    }
  }

  tx.Commit()

  return nil
}

func UpdateAsset(asset *Asset) error {
  dbConn := GetDBConnection()
  defer dbConn.Close()

  tx, err := dbConn.Begin()
  if err != nil {
    return err
  }

  stmt, err := tx.Prepare("update assets set name = ?, isimage = ? where id = ?")
  if err != nil {
    return err
  }
  defer stmt.Close()

  _, err = stmt.Exec(asset.Name, asset.IsImage, &asset.Id)
  if err != nil {
    return err
  }

  tx.Commit()
  return nil
}

func ControllerShowIndex(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)
  error := false
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

func ControllerNewAsset(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)
  error := false

  tmpl, _ := template.ParseFiles("views/assets/new.html")

  err := tmpl.Execute(w, nil)
  if err != nil {
    fmt.Printf("Unable to write new template to response.\n")
    error = true
  }

  if error {
    http.ServeFile(w, r, "views/error.html")
  }
}

func ControllerCreateAsset(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)

  if r.Method == "POST" {
    err := r.ParseForm()
    if err != nil {
      // ERROR: Unable to read form data.
      fmt.Printf("ERROR: Bad form data\n")
      http.Redirect(w, r, "/", 302)
    } else {
      // err := formDecoder.Decode(asset, r.PostForm)
      // if err == nil {
      _, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
      if err == nil {
        mr := multipart.NewReader(r.Body, params["boundary"])
        form, err := mr.ReadForm(1024)
        if err == nil {
          asset := new(Asset)
          asset.Name = strings.Join(form.Value["name"], "")
          if strings.Join(form.Value["isimage"],"") == "on" {
            asset.IsImage = true
          } else {
            asset.IsImage = false
          }
          if err == nil {
            if len(form.File["file"]) == 1 {
              fileName := form.File["file"][0].Filename
              destFileName := os.Getenv("ASSET_REPO_UPLOAD_DIR") + "/" + url.QueryEscape(fileName)
              asset.FileName = destFileName
              asset.URL = os.Getenv("ASSET_REPO_BASE_URL") + "/" + fileName
              // os.Copy(form.File["file"][0].tmpfile, destFileName)
              tmpFile, err := form.File["file"][0].Open()
              if err == nil {
                if err == nil {
                  data, err := ioutil.ReadAll(tmpFile)
                  if err == nil {
                    err = ioutil.WriteFile(destFileName, data, 0660)
                    if err == nil {
                      err = CreateAsset(asset)
                    }
                    tmpFile.Close()
                  }
                }
              }
            }

            form.RemoveAll()
          }
        }
      } else {
        fmt.Printf("%s\n", r.PostForm)
        fmt.Printf("%s\n", err)
      }
    }
  }

  http.Redirect(w, r, "/assets", 302)
}

func ControllerEditAsset(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)
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
  auth.CheckAuthCookie(w, r)
  vars := mux.Vars(r)
  id, _ := strconv.ParseInt(vars["id"], 10, 32)

  _ = DestroyAsset(id)
  http.Redirect(w, r, "/assets", 302)
}

func ControllerUpdateAsset(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)
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
      err := formDecoder.Decode(asset, r.PostForm)
      if err == nil {
        asset.Id = &id
        _ = UpdateAsset(asset)
      } else {
        fmt.Printf("%s\n", r.PostForm)
        fmt.Printf("%s\n", err)
      }

      http.Redirect(w, r, "/assets", 302)
    }
  } else {
    http.Redirect(w, r, "/assets", 302)
  }
}
