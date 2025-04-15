package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/hello", hello)

	server := http.Server{
		Addr:    ":9080",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}

}

func index(w http.ResponseWriter, r *http.Request) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		w.Write([]byte("error"))
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		_ = conn.Close()
		return
	}
}
func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World from server"))

}
