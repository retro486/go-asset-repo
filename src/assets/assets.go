package assets

import (
  "auth"
  "net/http"
)

func ShowIndex(w http.ResponseWriter, r *http.Request) {
  auth.CheckAuthCookie(w, r)
  http.ServeFile(w, r, "views/assets/index.html")
}
