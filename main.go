// Ability to accept JSON requests with BASE64 encoded images.
// Ability to upload image by its URL (hosted somewhere in Internet).
// Create thumb square preview (100 x 100 px) for every uploaded image.
// The following will be appreciated:

// Graceful shutdown of application.
// Dockerfile and docker-compose.yml which allow to boot up application in a single docker-compose up command.
// Unit tests, functional tests, CI integration (Travis CI, Circle CI, etc).

package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"mime/multipart"

	"strings"

	"github.com/gorilla/mux"
)

const (
	maxUploadSize = 2 * 1024 // 2 mb
	imagesDir     = "images"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", upload)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}
	log.Print("Server started on localhost:8080, use /upload for uploading files")
	server.ListenAndServe()
}

// html upload page
func display(writer http.ResponseWriter, data interface{}, filename string) {
	file := fmt.Sprintf("%s.html", filename)
	t := template.Must(template.ParseFiles(file))
	t.ExecuteTemplate(writer, "layout", data)
}

func upload(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "Upload files", "index")

		//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		contentType := r.Header.Get("Content-Type")
		// multipart data upload
		if strings.HasPrefix(contentType, "multipart/form-data") {
			uploadMultipart(w, r)
		} else if contentType == "application/x-www-form-urlencoded" {
			//uploadByLink(w,r)
		} else {
			fmt.Println(contentType)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func uploadMultipart(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mf := r.MultipartForm
	//get the *fileheaders
	files := mf.File["files"]
	for _, file := range files {
		//for each fileheader, get a handle to the actual file
		err := saveFile(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	////display success message.
	display(w, "Success!", "index")
}

func saveFile(fh *multipart.FileHeader) error {
	mimeType := detectContentType(fh)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" {
		return fmt.Errorf("file should be an image: got %s", mimeType)
	}
	file, err := fh.Open()
	defer file.Close()
	if err != nil {
		return err
	}
	//create destination file making sure the path is write able.
	filename := fmt.Sprintf("%s/%s", imagesDir, fh.Filename)
	dst, err := os.Create(filename)
	defer dst.Close()
	if err != nil {
		return err
	}
	//copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return err
	}
	return nil
}

func detectContentType(fh *multipart.FileHeader) string {
	return fh.Header.Get("Content-Type")
}
