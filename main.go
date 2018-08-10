package main

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {

	setLogDirectory()

	f, err := os.OpenFile("logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		logError(err)
	}

	//
	http.HandleFunc("/", rootHandler)

	// upload file
	http.HandleFunc("/upload", uploadFileHandler)

	// file server
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/logs/", fs)

	log.SetOutput(f)
	logInfo("Server started on localhost:8888, use /upload for uploading files and /logs/{fileName} for downloading")

	logError(http.ListenAndServe(":8888", nil))

	defer f.Close()

}

func setLogDirectory() {
	logFilesPath := "logs"

	if _, err := os.Stat(logFilesPath); os.IsNotExist(err) {
		os.Mkdir(logFilesPath, 0777)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/upload", 301)
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles(filepath.Join("templates", "upload.tmpl"))
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		// fmt.Println("No memory problem")
		var err error
		for _, fheaders := range r.MultipartForm.File {
			for _, hdr := range fheaders {
				// open uploaded
				var infile multipart.File
				if infile, err = hdr.Open(); nil != err {
					logError(err)
					return
				}
				// open destination
				var outfile *os.File
				if outfile, err = os.Create("./logs/" + hdr.Filename); nil != err {
					logError(err)
					return
				}
				// 32K buffer copy
				var written int64
				if written, err = io.Copy(outfile, infile); nil != err {
					logError(err)
					return
				}
				w.Write([]byte("uploaded file:" + hdr.Filename + ";length:" + strconv.Itoa(int(written))))
			}
		}
	}

}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func logError(err interface{}) {
	fmt.Println(err)
	log.Fatalf("%v", err)
}

func logInfo(info interface{}) {
	fmt.Println(info)
	log.Printf("%v", info)
}
