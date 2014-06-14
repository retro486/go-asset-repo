// OTP and base server heavily based on https://github.com/jstoja/go-otp-server
package main

import (
  "auth"
  "github.com/gorilla/mux"
  "fmt"
  "log"
  "net/http"
)

func ShowIndex(w http.ResponseWriter, r *http.Request) {
  // Check session key authorized, if not redirect to /
  http.ServeFile(w, r, "views/index.html")
}

func ShowManage(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)
  http.ServeFile(w, r, "views/manage.html")
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", ShowIndex)
  r.HandleFunc("/manage", ShowManage)
  r.HandleFunc("/requireOTP", auth.HandleOTP)
  r.HandleFunc("/logout", auth.Logout)
  http.Handle("/", r)

  fmt.Printf("Listening on 0.0.0.0:8080...\n")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
