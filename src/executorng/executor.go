package executorng

import (
	"fmt"
	"service"
	"time"
)

type messageContainer struct {
	operation string // "immediate", "deferred", "done"
}

// A concurrent service executor
// Collection and rule handlers work as concurrent processes
// managed by the control loop in this function.
func Execute(s *service.Service) {
	// launch the handlers
	channels := makeChannels(s)
	for collectionName, _ := range s.Collections {
		go collectionHandler(collectionName, s, channels)
	}
	for ruleNumber, _ := range s.Rules {
		go ruleHandler(ruleNumber, s, channels)
	}
	go budCommunicator(s, channels)

	// make list of all processes to be controlled
	toControl := []chan<- messageContainer{channels.distribution}
	for _, collectionChannel := range channels.collections {
		toControl = append(toControl, collectionChannel)
	}
	for _, ruleChannel := range channels.rules {
		toControl = append(toControl, ruleChannel)
	}

	fmt.Println(s)
	// control loop
	for {
		// time.Sleep(2*time.Second)
		startTime := time.Now()

		// phase 1: execute immediate rules
		sendAndWaitTilFinished(
			messageContainer{operation: "immediate"},
			toControl, channels.finished)

		// phase 2: execute deferred rules
		sendAndWaitTilFinished(
			messageContainer{operation: "deferred"},
			toControl, channels.finished)

		timeinfo("control", "timestep took", time.Since(startTime))
	}
}

func sendAndWaitTilFinished(message messageContainer, toControl []chan<- messageContainer, finishedChannel <-chan bool) {
	for _, channel := range toControl {
		channel <- message
	}
	for finished := 0; finished < len(toControl); finished++ {
		<-finishedChannel
		controlinfo("control", finished, "of", len(toControl))
	}
}
