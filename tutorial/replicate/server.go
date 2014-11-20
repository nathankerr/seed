package main

import (
	"fmt"
	"net/http"
)

const ADDR = ":4000"

func main() {
	http.Handle("/", http.FileServer(http.Dir(".")))

	fmt.Println("Starting Put client on", ADDR)
	err := http.ListenAndServe(ADDR, nil)
	if err != nil {
		println(err.Error())
	}
}
