package main

import(
	"net/http"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))

	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		println(err.Error())
	}
}