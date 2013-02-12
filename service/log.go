package service

// control output verbosity by toggling lexinfo, parseinfo, and info on and off
// by enabling/disabling the return statements.

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func lexinfo(args ...interface{}) {
	return
	printlog(args...)
}

func parseinfo(args ...interface{}) {
	return
	printlog(args...)
}

func info(args ...interface{}) {
	return
	printlog(args...)
}

func fatal(args ...interface{}) {
	printlog(args...)
	os.Exit(1)
}

func fatalf(format string, args ...interface{}) {
	printlog(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// include source location with log output
func printlog(args ...interface{}) {
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

	fmt.Printf("%s", info)
}
