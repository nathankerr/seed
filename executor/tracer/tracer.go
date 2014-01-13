// a monitor to output execution traces
package tracer

import (
	"encoding/json"
	"github.com/nathankerr/seed/executor"
	"log"
	"os"
)

func StartTracer(filename string, monitor <-chan executor.MonitorMessage) {
	traceFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	trace := log.New(traceFile, "", log.Lmicroseconds)

	for {
		message := <-monitor

		encoded, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
			continue
		}
		trace.Printf("%s\n", string(encoded))
	}
}
