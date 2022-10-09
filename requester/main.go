package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
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
	"/choco/":           "https://community.chocolatey.org/api/v2/",
	"/debian/":          "http://ftp.jp.debian.org/debian/",
	"/debian-security/": "http://security.debian.org/debian-security/",
	"/proxmox/":         "http://download.proxmox.com/debian/",
	"/arch/":            "http://ftp.jaist.ac.jp/pub/Linux/ArchLinux/",
	"/any/":             "",
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
	proto := "http"
	if r.Header.Get("X-Forwarded-Proto") != "" {
		proto = r.Header.Get("X-Forwarded-Proto")
	}
	host := r.Host
	if r.Header.Get("X-Forwarded-Host") != "" {
		host = r.Header.Get("X-Forwarded-Host")
	}
	for k, v := range convert {
		if strings.HasPrefix(r.RequestURI, k) {
			dest = strings.Replace(r.RequestURI, k, v, 1)
			sentFrom = proto + "://" + host + k
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
	if strings.Contains(res.Header.Get("content-type"), "application/atom+xml") {
		var buf = &bytes.Buffer{}
		io.Copy(buf, res.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("PCCCache failed to read responce: " + err.Error()))
			log.Print("Failed to read responce: ", err)
			return
		}
		str := buf.String()
		buf.Reset()
		str = strings.ReplaceAll(str, sendTo, sentFrom)
		if strings.HasPrefix(sendTo, "https://") {
			str = strings.ReplaceAll(str, strings.Replace(sendTo, "https://", "http://", 1), sentFrom)
		}
		if strings.HasPrefix(sendTo, "http://") {
			str = strings.ReplaceAll(str, strings.Replace(sendTo, "http://", "https://", 1), sentFrom)
		}
		buf.WriteString(str)
		w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		w.WriteHeader(res.StatusCode)
		io.Copy(w, buf)
	} else {
		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	}
}
