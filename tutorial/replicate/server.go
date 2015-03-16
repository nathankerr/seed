package main

import (
	"fmt"
	"net/http"
)

const addr = ":4000"

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Starting Put client on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		println(err.Error())
	}
}
