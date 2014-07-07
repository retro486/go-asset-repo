package auth

import (
  "github.com/retro486/go-asset-repo/otp"
  "github.com/gorilla/schema"
  "github.com/gorilla/securecookie"
  "log"
  "net/http"
  "os"
  "strconv"
)

type OTPForm struct {
  OTP string `schema:"otp"`
}

var HMACKey = GetHMAC() // or GenerateRandomKey()
var secureCookie = securecookie.New(HMACKey, nil)

func GetPassword() string {
  return os.Getenv("ASSET_REPO_OTP")
}

func GetHMAC() []byte {
  return []byte(os.Getenv("ASSET_REPO_HMAC"))
}

func SetAuthCookie(w http.ResponseWriter, r *http.Request) {
  value := map[string]bool {
    "authorized": true,
  }

  if encoded, err := secureCookie.Encode("auth", value); err == nil {
    cookie := &http.Cookie{
      Name: "auth",
      Value: encoded,
      Path: "/",
    }
    http.SetCookie(w, cookie)
  }
}

func ClearAuthCookie(w http.ResponseWriter, r *http.Request) {
  cookie := &http.Cookie {
    Name: "auth",
    MaxAge: -1,
  }
  http.SetCookie(w, cookie)
}

func CheckAuthCookie(w http.ResponseWriter, r *http.Request) {
  if cookie, err := r.Cookie("auth"); err == nil {
    value := make(map[string]bool)
    if err = secureCookie.Decode("auth", cookie.Value, &value); err == nil {
      if value["authorized"] != true {
        http.Redirect(w, r, "/", 302)
      }
    }
  } else {
    // cookie doesn't exist
    http.Redirect(w, r, "/", 302)
  }
}

func ControllerLogout(w http.ResponseWriter, r *http.Request) {
  ClearAuthCookie(w, r)
  http.Redirect(w, r, "/", 302)
}

func ControllerLogin(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    err := r.ParseForm()
    if err != nil {
      // ERROR: Unable to read form data.
      log.Print("ERROR: Bad form data\n")
      http.Redirect(w, r, "/", 302)
    } else {
      otp_form := new(OTPForm)
      decoder := schema.NewDecoder()
      err := decoder.Decode(otp_form, r.PostForm)
      cmp := gotp.Totp(GetPassword())

      if err != nil {
        log.Print("ERROR: Can't decode form\n")
      } else {
        str_i, _ := strconv.ParseUint(otp_form.OTP, 10, 32)
        if cmp == uint32(str_i) {
          log.Print("OTP Authorized\n")

          // Set session key for authorized
          SetAuthCookie(w, r)
          http.Redirect(w, r, "/assets", 302)
        } else {
          // ERROR: Invalid OTP value given.
          log.Print("ERROR: Bad OTP\n")
          http.Redirect(w, r, "/", 302)
        }
      }
    }
  }
}
