package executorng

import (
	"service"
)

func ruleHandler(ruleNumber int, s *service.Service, channels channels) {
	controlinfo(ruleNumber, "started")
	input := channels.rules[ruleNumber]
	rule := s.Rules[ruleNumber]
	dataMessages := []messageContainer{}

	for {
		message := <-input
		controlinfo(ruleNumber, "received", message)

		switch message.operation {
		case "immediate":
			if rule.Operation == "<=" {
				runRule(ruleNumber, s, channels, dataMessages)
			}
			dataMessages = []messageContainer{}
			channels.finished <- true
			controlinfo(ruleNumber, "finished with", message)
		case "deferred":
			if rule.Operation != "<=" {
				runRule(ruleNumber, s, channels, dataMessages)
			}
			dataMessages = []messageContainer{}
			channels.finished <- true
			controlinfo(ruleNumber, "finished with", message)
		case "data":
			dataMessages = append(dataMessages, message)
		default:
			fatal(ruleNumber, "unhandled message:", message)
		}
	}
}

func runRule(ruleNumber int, s *service.Service, channels channels, dataMessages []messageContainer) {
	rule := s.Rules[ruleNumber]
	requires := rule.Requires()
	input := channels.rules[ruleNumber]

	outputName := s.Rules[ruleNumber].Supplies
	output := channels.collections[outputName]
	controlinfo(ruleNumber, "sends to", outputName)
	controlinfo(ruleNumber, "requires", requires)

	//TODO handle early recieved dataMessages

	stillRequired := len(requires) - len(dataMessages)

	for i := 0; i < stillRequired; i++ {
		message := <-input
		controlinfo(ruleNumber, "received", message)

		switch message.operation {
		case "data":
			// TODO: store data
		default:
			fatal(ruleNumber, "unhandled message", message)
		}
	}

	resultMessage := messageContainer{
		operation:  "data",
		collection: outputName,
		data:       []tuple{},
	}
	controlinfo(ruleNumber, "sending", resultMessage)
	output <- resultMessage
}
