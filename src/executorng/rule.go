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
			// cache data received before an immediate or deferred message initiates execution
			dataMessages = append(dataMessages, message)
		default:
			fatal(ruleNumber, "unhandled message:", message)
		}
	}
}

func runRule(ruleNumber int, s *service.Service, channels channels, dataMessages []messageContainer) {
	// get the data needed to calculate the results
	_ = getRequiredData(ruleNumber, s.Rules[ruleNumber], dataMessages, channels.rules[ruleNumber])

	// calculate results
	results := []tuple{}

	// send results
	outputName := s.Rules[ruleNumber].Supplies
	channels.collections[outputName] <- messageContainer{
		operation:  "data",
		collection: outputName,
		data:       results,
	}
}

func getRequiredData(ruleNumber int, rule *service.Rule, dataMessages []messageContainer, input <-chan messageContainer) map[string][]tuple {
	data := map[string][]tuple{}

	// process cached data
	for _, message := range dataMessages {
		data[message.collection] = message.data
	}

	// receive other needed data
	for stillNeeded := len(rule.Requires()) - len(dataMessages); stillNeeded > 0; stillNeeded-- {
		message := <-input
		controlinfo(ruleNumber, "received", message)

		switch message.operation {
		case "data":
			data[message.collection] = message.data
		default:
			fatal(ruleNumber, "unhandled message", message)
		}
	}

	return data
}
