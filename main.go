package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type M map[string]interface{}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/save", handleSave)
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/list-files", handleListFiles)
	http.HandleFunc("/download", handleDownloadView)
	http.HandleFunc("/download/get", handleDownloadGet)

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("assets"))))

	fmt.Println("server started at localhost:9000")
	http.ListenAndServe(":9000", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("view.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSave(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		payload := struct {
			Name string `json:"name"`
			Age int `json:"age"`
			Gender string `json:"gender"`
		}{}
		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.Error(w, "Only accept POST requet", http.StatusBadRequest)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("upload.html"))
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "POST" {
		basePath, _ := os.Getwd()
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			fileLocation := filepath.Join(basePath, "files", part.FileName())
			dst, err := os.Create(fileLocation)

			if dst != nil {
				defer dst.Close()
			}

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Write([]byte("all files uploaded"))
		return
	}
}

func handleDownloadView(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("download.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	files := []M{}
	basePath, _ := os.Getwd()
	filesLocation := filepath.Join(basePath, "files")

	err := filepath.Walk(filesLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, M{"filename": info.Name(), "path": path})
		return nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func handleDownloadGet(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	path := r.FormValue("path")
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	contentDisposition := fmt.Sprintf("attachment; filename=%s", f.Name())
	w.Header().Set("Content-Disposition", contentDisposition)

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
