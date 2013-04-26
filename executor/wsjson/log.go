package wsjson

// control output verbosity by toggling lexinfo, parseinfo, and info on and off
// by enabling/disabling the return statements.

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

func info(args ...interface{}) {
	return
	printlog("wsjson", args...)
}

func log(args ...interface{}) {
	// return
	printlog("wsjson", args...)
}

func fatal(id interface{}, args ...interface{}) {
	info := ""

	pc, file, line, ok := runtime.Caller(1)
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

	info += fmt.Sprintf("[%v] ", id)
	info += fmt.Sprintln(args...)

	fmt.Print(info)
	os.Exit(1)
}

// include source location with log output
func printlog(id interface{}, args ...interface{}) {
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

	info = fmt.Sprintf("%s[%v] ", info, id)
	info += fmt.Sprintln(args...)

	fmt.Printf("%s", info)
}
