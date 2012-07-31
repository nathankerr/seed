package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	kvs, err := ioutil.ReadFile("kvs.seed")
	if err != nil {
		panic(err)
	}

	seed := parse("kvs.seed", string(kvs))

	fmt.Println(seed)
}
