package imagedatastore

import (
    "html/template"
    "io"
    "net/http"
    "appengine"
    "appengine/datastore"
    "appengine/blobstore"
    "fmt"
)

type Image struct {
    BlobKey appengine.BlobKey
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    datastoreKey := r.URL.Path[1:]
    if len(datastoreKey) > 0 {
        var img Image

        key, err := datastore.DecodeKey(datastoreKey)
        if err == nil {
            if err = datastore.Get(c, key, &img); err == nil {
                blobstore.Send(w, img.BlobKey)
                return
            }
        }
    }

    fmt.Fprint(w, "Hello World!")
}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
    w.WriteHeader(http.StatusInternalServerError)
    w.Header().Set("Content-Type", "text/plain")
    io.WriteString(w, "Internal Server Error")
    c.Errorf("%v", err)
}

var uploadTemplate = template.Must(template.New("root").Parse(uploadTemplateHTML))

const uploadTemplateHTML = `
<html><body>
<form action="{{.}}" method="POST" enctype="multipart/form-data">
Upload File: <input type="file" name="file"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
`

func handleUpload(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    uploadURL, err := blobstore.UploadURL(c, "/doupload", nil)
    if err != nil {
        serveError(c, w, err)
        return
    }
    w.Header().Set("Content-Type", "text/html")
    err = uploadTemplate.Execute(w, uploadURL)
    if err != nil {
        c.Errorf("%v", err)
    }
}

func handleDoUpload(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    blobs, _, err := blobstore.ParseUpload(r)
    if err != nil {
        serveError(c, w, err)
        return
    }
    file := blobs["file"]
    if len(file) == 0 {
        c.Errorf("no file uploaded")
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    img := Image {
        BlobKey: file[0].BlobKey,
    }

    key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "image", nil), &img)
    if err != nil {
        c.Errorf("datastore fail")
        http.Redirect(w, r, "/", http.StatusFound)
        return
    }

    http.Redirect(w, r, "/"+key.Encode(), http.StatusFound)
}

func init() {
    http.HandleFunc("/", handleRoot)
    http.HandleFunc("/upload", handleUpload)
    http.HandleFunc("/doupload", handleDoUpload)
}
