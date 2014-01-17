package monitor

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// control output verbosity by toggling the following constants
const (
	logMONITORINFO  = false
	logNETWORKERROR = true
	logFLOWINFO     = false
	logINFO         = false
	logTIMEINFO     = false
	logCONTROLINFO  = false
)

// centralized monitoring info
func monitorinfo(id interface{}, args ...interface{}) {
	if logMONITORINFO {
		printlog(id, args...)
	}
}

func networkerror(id interface{}, args ...interface{}) {
	if logNETWORKERROR {
		printlog(id, args...)
	}
}

// data sent and received between go routines
func flowinfo(id interface{}, args ...interface{}) {
	if logFLOWINFO {
		printlog(id, args...)
	}
}

func info(id interface{}, args ...interface{}) {
	if logINFO {
		printlog(id, args...)
	}
}

// timing information from executor
func timeinfo(id interface{}, args ...interface{}) {
	if logTIMEINFO {
		printlog(id, args...)
	}
}

// control communication information
func controlinfo(id interface{}, args ...interface{}) {
	if logCONTROLINFO {
		printlog(id, args...)
	}
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
	info := fmt.Sprintf("[%v] ", id)
	info += fmt.Sprintln(args...)

	fmt.Printf("%s", info)
}
