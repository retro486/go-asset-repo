### Go OTP Asset Repository
I needed a simple asset repo to upload static assets for use on my blog and I wanted to refresh my Google Go since I hadn't touched it since the beta days (as seen by my now complete non-working go-dungeon project).

Two birds, one stone.

Environment variables required:

* GOPATH - the top level project folder (where this README is located)
* ASSET_REPO_TEMPLATES - The path where HTML templates are located
* ASSET_REPO_UPLOAD_DIR - The path to upload files to
* ASSET_REPO_BASE_URL - The URL where upload dir is accessible
* ASSET_REPO_OTP - Your OTP secret
* ASSET_REPO_HMAC - Your cookie secret
