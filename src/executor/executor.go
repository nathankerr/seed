package executor

import (
	"fmt"
	"service"
	"strings"
	"time"
)

type tuple []interface{}
type messageContainer struct {
	operation  string // "immediate", "deferred", "data", "done"
	collection string
	data       []tuple
}

func (tuple *tuple) String() string {
	columns := []string{}
	for _, column := range *tuple {
		switch typed := column.(type) {
		case []byte:
			columns = append(columns, string(typed))
		default:
			columns = append(columns, fmt.Sprintf("%#v", column))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(columns, ", "))
}

func (mc *messageContainer) String() string {
	tuples := []string{}
	for _, tuple := range mc.data {
		tuples = append(tuples, tuple.String())
	}
	return fmt.Sprintf("{%s %s [%s]}",
		mc.collection, mc.operation, strings.Join(tuples, ", "))
}

// A concurrent service executor
// Collection and rule handlers work as concurrent processes
// managed by the control loop in this function.
func Execute(s *service.Service, timeoutDuration time.Duration, sleepDuration time.Duration, address string, monitorAddress string) {
	// launch the handlers
	channels := makeChannels(s)
	for collectionName, _ := range s.Collections {
		go collectionHandler(collectionName, s, channels)
	}
	for ruleNumber, _ := range s.Rules {
		go handleRule(ruleNumber, s, channels)
	}
	go budCommunicator(s, channels, address)

	// make list of all processes to be controlled
	toControl := []chan<- messageContainer{channels.distribution}
	for _, collectionChannel := range channels.collections {
		toControl = append(toControl, collectionChannel)
	}
	for _, ruleChannel := range channels.rules {
		toControl = append(toControl, ruleChannel)
	}

	info("execute", "service", s)
	info("execute", "channels", channels)

	monitor := false
	// var monitorChannel chan monitorMessage
	monitorChannel := make(chan monitorMessage)
	if monitorAddress != "" {
		monitor = true
		go startMonitor(monitorAddress, monitorChannel)
	}

	// setup and start the timeout
	// timeout should only happen when timeoutDuration != 0
	var timeout <-chan time.Time
	if timeoutDuration != 0 {
		timeout = time.After(timeoutDuration)
	}
	// control loop
	for {
		breakBeforeDeferred := false
		if monitor {
			// message = <-monitorChannel
		}

		startTime := time.Now()
		time.Sleep(sleepDuration)

		// check for timeout
		select {
		case <-timeout:
			info("execute", "Timeout")
			return
		default:
		}

		// phase 1: execute immediate rules
		sendAndWaitTilFinished(
			messageContainer{operation: "immediate"},
			toControl, channels.control)

		if breakBeforeDeferred {
			// handle monitor control
		}

		// phase 2: execute deferred rules
		sendAndWaitTilFinished(
			messageContainer{operation: "deferred"},
			toControl, channels.control)

		timeinfo("execute", "timestep took", time.Since(startTime))
		if monitor {
			monitorChannel <- monitorMessage{operation: "time", data: time.Since(startTime)}
		}
	}
}

func sendAndWaitTilFinished(message messageContainer, toControl []chan<- messageContainer, controlChannel <-chan messageContainer) {
	sendToAll(message, toControl)
	for finished := 0; finished < len(toControl); finished++ {
		message := <-controlChannel
		monitorinfo("MONITOR", message.String())
		controlinfo("execute", finished, "of", len(toControl))
	}
}

func sendToAll(message messageContainer, channels []chan<- messageContainer) {
	for _, channel := range channels {
		channel <- message
	}
}
