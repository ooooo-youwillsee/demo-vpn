package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)

	server := http.Server{
		Addr:    ":7080",
		Handler: mux,
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}

}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World from APP"))
}
