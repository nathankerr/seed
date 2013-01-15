package executorng

import (
	"service"
)

func collectionHandler(collectionName string, s *service.Service, channels channels) {
	controlinfo(collectionName, "started")
	input := channels.collections[collectionName]
	c := s.Collections[collectionName]

	immediates := ruleChannels(false, collectionName, s, channels)
	deferreds := ruleChannels(true, collectionName, s, channels)

	controlinfo(collectionName, "sends to", immediates, deferreds)

	// data := tupleSet{}

	for {
		message := <-input
		controlinfo(collectionName, "received", message)

		switch message.operation {
		case "immediate":
			// info(collectionName, "immediate")
			sendToAll(messageContainer{operation: "data"}, immediates)
			channels.finished <- true
			controlinfo(collectionName, "finished with", message)
		case "deferred":
			// info(collectionName, "deferred")
			controlinfo(collectionName, "sending to", deferreds)
			sendToAll(messageContainer{collection: collectionName, operation: "data"}, deferreds)
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
			channels.finished <- true
			controlinfo(collectionName, "finished with", message)
		case "data":
			//TODO
		default:
			fatal(collectionName, "unhandled message:", message)
		}

	}
}

func ruleChannels(deferred bool, collectionName string, s *service.Service, channels channels) []chan<- messageContainer {
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
		operation:  "data",
		collection: name,
		data:       []tuple{},
	}

	for _, tuple := range collection {
		message.data = append(message.data, tuple)
	}

	return message
}
