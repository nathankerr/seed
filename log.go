package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
)

// toggle return to turn on and off
func transformationinfo(args ...interface{}) {
	return
	info(args...)
}

func info(args ...interface{}) {
	info := ""

	pc, file, line, ok := runtime.Caller(2)
	if ok {
		basepath, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		sourcepath, err := filepath.Rel(basepath, file)
		if err != nil {
			panic(err)
		}
		info += fmt.Sprintf("%s:%d: ", sourcepath, line)

		name := path.Ext(runtime.FuncForPC(pc).Name())
		info += name[1:]
		if len(args) > 0 {
			info += ": "
		}
	}
	info += fmt.Sprintln(args...)

	log.Print(info)
}
