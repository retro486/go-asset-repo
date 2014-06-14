// OTP and base server heavily based on https://github.com/jstoja/go-otp-server
package main

import (
  "auth"
  "assets"
  "github.com/gorilla/mux"
  "fmt"
  "log"
  "net/http"
)

func ControllerShowIndex(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, "views/index.html")
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", ControllerShowIndex)

  r.HandleFunc("/login", auth.ControllerLogin).Methods("POST")
  r.HandleFunc("/logout", auth.ControllerLogout)

  r.HandleFunc("/assets", assets.ControllerShowIndex)
  // r.HandleFunc("/assets/new", assets.ControllerNewAsset)
  r.HandleFunc("/assets/{id}/destroy", assets.ControllerDestroyAsset)
  // r.HandleFunc("/assets/{id}/edit", assets.ControllerEditAsset)

  http.Handle("/", r)

  fmt.Printf("Listening on 0.0.0.0:8080...\n")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
