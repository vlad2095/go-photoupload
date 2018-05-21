// Ability to accept multiple files.
// Ability to accept multipart/form-data requests.
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
	"net/http"

	"github.com/gorilla/mux"
	"io"
	"log"
	"os"
)

const maxUploadSize = 2 * 1024 // 2 mb
const imagesDir = "images"

func main() {
	router := mux.NewRouter()

	fs := http.FileServer(http.Dir(imagesDir))
	router.Handle("/files/", http.StripPrefix("/files", fs))
	router.HandleFunc("/", handler)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}
	log.Print("Server started on localhost:8080, use /upload for uploading files")
	server.ListenAndServe()
}

func display(writer http.ResponseWriter, data interface{}, filename string) {
	file := fmt.Sprintf("%s.html", filename)
	templates := template.Must(template.ParseFiles(file))
	templates.ExecuteTemplate(writer, "layout", data)
}

//This is where the action happens.
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "Add files", "index")

		//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		//parse the multipart form in the request
		err := r.ParseMultipartForm(maxUploadSize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//get a ref to the parsed multipart form
		m := r.MultipartForm

		//get the *fileheaders
		files := m.File["files"]
		for i := range files {
			//for each fileheader, get a handle to the actual file
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//create destination file making sure the path is writeable.
			filename := fmt.Sprintf("%s/%s", imagesDir, files[i].Filename)
			dst, err := os.Create(filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//copy the uploaded file to the destination file
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}
		//display success message.
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
