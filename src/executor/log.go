package executor

// control output verbosity by toggling lexinfo, parseinfo, and info on and off
// by enabling/disabling the return statements.

import (
	"fmt"
	// "os"
	// "path"
	// "path/filepath"
	// "runtime"
	"time"
)

func operationinfo(id interface{}, args ...interface{}) {
	return
	printlog(id, args...)
}

func monitorinfo(id interface{}, args ...interface{}) {
	return
	printlog(id, args...)
}

func flowinfo(id interface{}, args ...interface{}) {
	return
	printlog(id, args...)
}

func info(id interface{}, args ...interface{}) {
	printlog(id, args...)
}

// include source location with log output
func printlog(id interface{}, args ...interface{}) {
	info := fmt.Sprintf("%v [%v] ", time.Now().Format("15:04:05.000000"), id)

	// pc, file, line, ok := runtime.Caller(2)
	// if ok {
	// 	basepath, err := filepath.Abs(".")
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	sourcepath, err := filepath.Rel(basepath, file)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	info += fmt.Sprintf("%s:%d: ", sourcepath, line)

	// 	name := path.Ext(runtime.FuncForPC(pc).Name())
	// 	info += name[1:]
	// 	if len(args) > 0 {
	// 		info += ": "
	// 	}
	// }
	info += fmt.Sprintln(args...)

	fmt.Printf("%s", info)
}
