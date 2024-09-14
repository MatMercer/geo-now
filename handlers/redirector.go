package handlers

import _ "embed"
import "net/http"

//go:embed redirector.html
var redirectorHtml []byte

// RedirectorHandler redirects the user to the device's browser resolution URL, like /goes/800x600
func RedirectorHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(redirectorHtml)
}
