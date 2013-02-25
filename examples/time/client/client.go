package main

import (
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))

	println("Starting time client on http://127.0.0.1:4000")
	err := http.ListenAndServe("127.0.0.1:4000", nil)
	if err != nil {
		println(err.Error())
	}
}
