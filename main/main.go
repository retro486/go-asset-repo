// OTP and base server heavily based on https://github.com/jstoja/go-otp-server
package main

import (
  "github.com/retro486/go-asset-repo/auth"
  "github.com/retro486/go-asset-repo/assets"
  "github.com/gorilla/mux"
  "fmt"
  "log"
  "net/http"
  "os"
)

func ControllerShowIndex(w http.ResponseWriter, r *http.Request) {
  templatePath := os.Getenv("ASSET_REPO_TEMPLATES")
  http.ServeFile(w, r, templatePath + "/index.html")
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", ControllerShowIndex)

  r.HandleFunc("/login", auth.ControllerLogin).Methods("POST")
  r.HandleFunc("/logout", auth.ControllerLogout).Methods("GET")

  r.HandleFunc("/assets", assets.ControllerShowIndex).Methods("GET")
  r.HandleFunc("/assets", assets.ControllerCreateAsset).Methods("POST")
  r.HandleFunc("/assets/new", assets.ControllerNewAsset).Methods("GET")
  // r.HandleFunc("/assets/{id}", assets.ControllerShowAsset).Methods("GET")
  r.HandleFunc("/assets/{id}/destroy", assets.ControllerDestroyAsset).Methods("GET")
  r.HandleFunc("/assets/{id}/edit", assets.ControllerEditAsset).Methods("GET")
  r.HandleFunc("/assets/{id}", assets.ControllerUpdateAsset).Methods("POST")

  http.Handle("/", r)

  fmt.Printf("Listening on 0.0.0.0:8080...\n")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
