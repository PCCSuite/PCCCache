package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func main() {
	server := &http.Server{
		Addr:    ":8080",
		Handler: &Handler{},
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Panic("Failed to listen and serve: ", err)
	}
}

type Handler struct {
}

var convert = map[string]string{
	"/choco/":  "https://community.chocolatey.org/api/v2/",
	"/debian/": "http://ftp.jp.debian.org/debian/",
	"/arch/":   "http://ftp.jaist.ac.jp/pub/Linux/ArchLinux/",
	"/any/":    "",
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(405)
		w.Write([]byte("Invalid method"))
		log.Print("Invalid method: ", r.Method)
		return
	}
	var dest string
	var sentFrom string
	var sendTo string
	for k, v := range convert {
		if strings.HasPrefix(r.RequestURI, k) {
			dest = strings.Replace(r.RequestURI, k, v, 1)
			sentFrom = "http://" + r.Host + k
			sendTo = v
			break
		}
	}
	if dest == "" {
		w.WriteHeader(400)
		w.Write([]byte("Invalid path"))
		log.Print("Invalid path: ", r.RequestURI)
		return
	}
	res, err := http.Get(dest)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("PCCCache failed to request: " + err.Error()))
		log.Print("Failed to read request: ", err)
		return
	}
	defer res.Body.Close()
	for k, v := range res.Header {
		for i, v2 := range v {
			if i == 0 {
				w.Header().Set(k, v2)
			} else {
				w.Header().Add(k, v2)
			}
		}
	}
	if res.Header.Get("content-type") == "application/atom+xml" {
		buf, err := io.ReadAll(res.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("PCCCache failed to read responce: " + err.Error()))
			log.Print("Failed to read responce: ", err)
			return
		}
		result := strings.ReplaceAll(string(buf), sendTo, sentFrom)
		w.WriteHeader(res.StatusCode)
		w.Write([]byte(result))
	} else {
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	}
}
