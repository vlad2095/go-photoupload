// Ability to accept JSON requests with BASE64 encoded images.
// Create thumb square preview (100 x 100 px) for every uploaded image.

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

	"strings"

	"github.com/disintegration/imaging"
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
			err := uploadMultipart(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else if contentType == "application/x-www-form-urlencoded" {
			err := uploadLink(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//} else if contentType == "application/json" {
			//	err := uploadJson(r)
			//	if err != nil {
			//		http.Error(w, err.Error(), http.StatusInternalServerError)
			//		return
			//	}
		} else {
			http.Error(w, "unsupported content-type: "+contentType, http.StatusInternalServerError)

		}
		display(w, "Success!", "index")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func uploadMultipart(r *http.Request) error {
	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		return err
	}
	form := r.MultipartForm
	//get the *fileheaders
	files := form.File["image"]
	for _, fh := range files {
		//for each fileheader, get a handle to the actual file
		mimeType := fh.Header.Get("Content-Type")
		if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" {
			return fmt.Errorf("file should be an image: got %s", mimeType)
		}
		file, err := fh.Open()
		if err != nil {
			return err
		}

		err = saveFile(file, fh.Filename)

		file.Close()

		if err != nil {
			return err
		}

		err = createThumbnail(fh.Filename)
		if err != nil {
			log.Println(err.Error())
		}
	}
	return nil
}

func uploadLink(r *http.Request) error {
	link := r.FormValue("image")
	resp, err := http.Get(link)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = saveFile(resp.Body, link[len(link)-10:])
	if err != nil {
		return err
	}

	err = createThumbnail(link[len(link)-10:])
	if err != nil {
		log.Println(err.Error())
	}
	return nil
}

func saveFile(file io.Reader, filename string) error {
	//create destination file making sure the path is write able.
	filename = fmt.Sprintf("%s/%s", imagesDir, filename)
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

func createThumbnail(filename string) error {
	path := fmt.Sprintf("%s/%s", imagesDir, filename)
	img, err := imaging.Open(path)
	if err != nil {
		return err
	}

	thumb := imaging.Thumbnail(img, 100, 100, imaging.CatmullRom)
	// save resized image
	err = imaging.Save(thumb, fmt.Sprintf("%s/thumb_%s", imagesDir, filename))

	if err != nil {
		return err
	}
	return nil
}
