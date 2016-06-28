package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

const SwaggerBaseDir = "/tmp/swagger-ui/"

type SwaggerHandler struct {
	GoboardHandler

	baseDir http.Dir
}

func NewSwaggerHandler() (s *SwaggerHandler) {
	s = &SwaggerHandler{}

	s.baseDir = http.Dir(SwaggerBaseDir)

	s.supportedOps = []SupportedOp{
		{"/swagger/", "GET"}, // GET swagger content
		{"/swagger/{file}", "GET"}, // GET swagger content
		{"/swagger/{subdir}/{file}", "GET"}, // GET swagger subdir content
	}
	return
}

func (s *SwaggerHandler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {

	vars := mux.Vars(rq)
	filePath := vars["file"]
	subDirPath := vars["subdir"]
	if len(subDirPath) > 0 {
		filePath = subDirPath + "/" + filePath
	}

	if len(filePath) == 0 || strings.HasSuffix(filePath, "/") {
		filePath = filePath + "index.html"
	}

	fmt.Println(filePath)

	if f, err := s.baseDir.Open(filePath); err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		return
	} else {
		defer f.Close()

		if fStat, err := f.Stat(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			http.ServeContent(w, rq, fStat.Name(), fStat.ModTime(), f)
		}
	}
	return
}