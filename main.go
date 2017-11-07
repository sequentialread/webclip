package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
  "strings"
  "io"
  "bytes"
)

var clipboard []byte
var filename string

func mainHandler(response http.ResponseWriter, request *http.Request) {

  if request.Method == "GET" {

    requestPath := strings.TrimPrefix(request.URL.Path, "/")
    if len(requestPath) > 0 {
      response.Header().Add("content-type", "text/plain")

      fmt.Fprintf(response, `
#!/bin/bash

function wrapper {
  filename="%s"

  if [ ! -f "$filename" ]; then
    echo "Error: $filename is not a file."
		exit 1
	fi

  curl -s -X POST -H "X-File-Name: $filename" -H "Content-Type: application/octet-stream" --data-binary "@$filename" https://%s/
}

wrapper

`, requestPath, request.Host)

    } else if filename != "" && len(clipboard) > 0 {
      response.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
      response.Header().Add("Content-Type", "application/octet-stream")
      io.Copy(response, bytes.NewBuffer(clipboard))
			clipboard = make([]byte, 0)
			filename = ""
    } else {
			response.Header().Add("content-type", "text/plain")
  		response.WriteHeader(500)
      fmt.Fprint(response, "404 nothing in webclip.\n")
		}
  }

  if request.Method == "POST" {
    bodyBytes, err := ioutil.ReadAll(request.Body)
  	if err != nil {
			response.Header().Add("content-type", "text/plain")
  		response.WriteHeader(500)
      fmt.Fprint(response, "500 internal server error: could not read body.\n")
  	} else {
      filename = request.Header.Get("X-File-Name")
      clipboard = bodyBytes
      response.WriteHeader(200)
      fmt.Fprintf(response, "200 ok I got \"%s\" with %d bytes.\n\n", filename, len(clipboard))
    }
  }

}

func main() {
	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":8080", nil)
}
