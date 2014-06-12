// OTP and base server heavily based on https://github.com/jstoja/go-otp-server
package main

import (
  "./otp"
  "fmt"
  "log"
  "net/http"
  "os"
  "strconv"
)

func GetPassword() string {
  return os.Getenv("ASSET_REPO_OTP")
}

func ShowIndex(w http.ResponseWriter, r *http.Request) {
  // TODO r.GetCookie("authorized").value == "true" http.Redirect(w, r, "/manage", 302)
  http.ServeFile(w, r, "views/index.html")
}

func ShowManage(w http.ResponseWriter, r *http.Request) {
  cookie := http.Cookie {
    Name: "authorized",
    Value: "true",
    Secure: true,
  }
  http.SetCookie(w, &cookie)
  http.ServeFile(w, r, "views/manage.html")
}

func HandleOTP(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    err := r.ParseForm()
    if err != nil {
      // Var
      http.Redirect(w, r, "/", 302)
    } else {
      str := r.PostFormValue("otp")
      cmp := gotp.Totp(GetPassword())

      str_i, _ := strconv.ParseUint(str, 10, 32)
      if cmp == uint32(str_i) {
        fmt.Printf("authorized\n")
        http.Redirect(w, r, "/manage", 302)
      } else {
        // Set some variable accessible by view...?
        http.Redirect(w, r, "/", 302)
      }
    }
  }
}

func main() {
  http.HandleFunc("/", ShowIndex)
  http.HandleFunc("/manage", ShowManage)
  http.HandleFunc("/requireOTP", HandleOTP)
  fmt.Printf("Listening on 0.0.0.0:8080...\n")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
