package main

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/cors"
)

var info = []byte(`
cproxy: a proxy to enable cross origin resource sharing

usage:

to proxy a request, provide the url you wish to proxy as part of the url query
string, for example

    ?url=https://api.github.com/

will enable cors for that url
`)

func corsHandler(w http.ResponseWriter, r *http.Request) {
	corsURL := r.URL.Query().Get("url")
	if corsURL == "" {
		w.Write(info)
		return
	}
	parsed, err := url.Parse(corsURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !parsed.IsAbs() {
		http.Error(w, "URL must be absolute\n", http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(r.Method, parsed.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for key, val := range res.Header {
		w.Header().Set(key, val[0])
	}
	io.Copy(w, res.Body)
}

func reverseProxyHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse("https://github.com")
	if err != nil {
		panic(err)
	}
	p := httputil.NewSingleHostReverseProxy(u)
	p.ServeHTTP(w, r)
}

func main() {
	mux := http.NewServeMux()
	c := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowCredentials: true,
	})

	mux.HandleFunc("/", corsHandler)
	mux.HandleFunc("/github/", reverseProxyHandler)
	http.ListenAndServe(":3000", c.Handler(mux))
}
