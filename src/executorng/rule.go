package executorng

import (
	"service"
)

func ruleHandler(ruleNumber int, s *service.Service, channels channels) {
	controlinfo(ruleNumber, "started")
	input := channels.rules[ruleNumber]

	for {
		message := <-input
		controlinfo(ruleNumber, "received", message)
		channels.finished <- true
		controlinfo(ruleNumber, "finished with", message)
	}
}
