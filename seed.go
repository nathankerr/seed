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

	l := lex("kvs.seed", string(kvs))

	go l.run()

Loop:
	for {
		select {
		case item := <-l.items:
			if item.typ == itemEOF {
				break Loop
			}
			fmt.Println(item)
		}
	}

	fmt.Println("Done")
}
