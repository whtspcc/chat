package server

import "net/http"

func RunServer(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}
