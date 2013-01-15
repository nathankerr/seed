package executorng

import (
	"service"
)

func collectionHandler(collectionName string, s *service.Service, channels channels) {
	controlinfo(collectionName, "started")
	input := channels.collections[collectionName]

	immediateRuleChannels := ruleChannels(false, collectionName, s, channels)
	deferredRuleChannels := ruleChannels(true, collectionName, s, channels)

	c := s.Collections[collectionName]

	data := tupleSet{}

	for {
		message := <-input
		controlinfo(collectionName, "received", message)

		switch message.operation {
		case "immediate":
			// sendToAll(messageContainer{}, immediateRuleChannels)
		case "deferred":
			// sendToAll(messageContainer{}, deferredRuleChannels)
			switch c.Type {
			case service.CollectionInput, service.CollectionOutput, service.CollectionScratch, service.CollectionChannel:
				// temporary collections
				// TODO: empty collection
			case service.CollectionTable:
				// persistent collections
				// no-op
			default:
				fatal(collectionName, "unhandled collection type", c.Type)
			}
		default:
			fatal(collectionName, "unhandled message:", message)
		}

		channels.finished <- true
		controlinfo(collectionName, "finished with", message)
	}
}

func ruleChannels (deferred bool, collectionName string, s *service.Service, channels channels) []chan<- messageContainer {
	ruleChannels := []chan<- messageContainer{}
	for ruleNum, rule := range s.Rules {
		if (deferred && rule.Operation == "<=") || (!deferred && rule.Operation != "<=") {
			continue
		}
		for _, collection := range rule.Requires() {
			if collection == collectionName {
				ruleChannels = append(ruleChannels, channels.rules[ruleNum])
				break
			}
		}
	}
	return ruleChannels
}

func collectionToMessage(name string, collection tupleSet) messageContainer {
	message := messageContainer{
		operation: "data",
		collection: name,
		data: tupleSet{},
	}

	for _, tuple := range collection {
		data.add(tuple)
	}
}