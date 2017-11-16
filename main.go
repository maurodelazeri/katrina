package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

var port = flag.Int("p", 8080, "Port number to listen to, defaults to 8080")
var uploadDir = flag.String("d", "/www", "Upload directory, defaults to '/www'")

// Request parameters
const (
	paramUuid      = "qquuid"      // uuid
	paramFile      = "qqfile"      // file name
	paramPartIndex = "qqpartindex" // part index

)

type UploadResponse struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	PreventRetry bool   `json:"preventRetry"`
}

func main() {
	flag.Parse()
	hostPort := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Initiating server listening at [%s]", hostPort)
	log.Printf("Base upload directory set to [%s]", *uploadDir)
	http.HandleFunc("/upload", UploadHandler)
	http.Handle("/upload/", http.StripPrefix("/www/", http.HandlerFunc(UploadHandler)))
	log.Fatal(http.ListenAndServe(hostPort, nil))
}

func UploadHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		upload(w, req)
		return
	case http.MethodDelete:
		delete(w, req)
		return
	}
	errorMsg := fmt.Sprintf("Method [%s] is not supported:", req.Method)
	http.Error(w, errorMsg, http.StatusMethodNotAllowed)
}

func upload(w http.ResponseWriter, req *http.Request) {
	uuid := req.FormValue(paramUuid)
	if len(uuid) == 0 {
		log.Printf("No uuid received, invalid upload request")
		http.Error(w, "No uuid received", http.StatusBadRequest)
		return
	}
	log.Printf("Starting upload handling of request with uuid of [%s]\n", uuid)
	file, headers, err := req.FormFile(paramFile)

	if err != nil {
		writeUploadResponse(w, err)
		return
	}

	fileDir := fmt.Sprintf("%s/%s", *uploadDir, uuid)
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		writeUploadResponse(w, err)
		return
	}

	var filename string
	partIndex := req.FormValue(paramPartIndex)
	if len(partIndex) == 0 {
		filename = fmt.Sprintf("%s/%s", fileDir, headers.Filename)

	} else {
		filename = fmt.Sprintf("%s/%s_%05s", fileDir, uuid, partIndex)
	}
	outfile, err := os.Create(filename)
	if err != nil {
		writeUploadResponse(w, err)
		return
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, file)
	if err != nil {
		writeUploadResponse(w, err)
		return
	}

	writeUploadResponse(w, nil)
}

func delete(w http.ResponseWriter, req *http.Request) {
	log.Printf("Delete request received for uuid [%s]", req.URL.Path)
	err := os.RemoveAll(*uploadDir + "/" + req.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)

}

func writeHttpResponse(w http.ResponseWriter, httpCode int, err error) {
	w.WriteHeader(httpCode)
	if err != nil {
		log.Printf("An error happened: %v", err)
		w.Write([]byte(err.Error()))
	}
}

func writeUploadResponse(w http.ResponseWriter, err error) {
	uploadResponse := new(UploadResponse)
	if err != nil {
		uploadResponse.Error = err.Error()
	} else {
		uploadResponse.Success = true
	}
	w.Header().Set("Content-Type", "text/plain")
	json.NewEncoder(w).Encode(uploadResponse)
}

func postFile(path, filename string, targetUrl string) error {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// this step is very important
	fileWriter, err := bodyWriter.CreateFormFile(paramFile, filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	// open file handle
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer fh.Close()

	//iocopy
	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return err
	}

	bodyWriter.WriteField(paramFile, filename)
	bodyWriter.WriteField(paramUuid, path)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}
