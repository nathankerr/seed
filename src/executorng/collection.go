package executorng

import (
	"encoding/json"
	"service"
)

type tupleSet struct {
	tuples          map[string]tuple
	keyEnds         int
	numberOfColumns int
	collectionName  string
}

// tuples are unique according to their key columns
// the key columns are a subset of the columns starting at the beginning
// encoding the key columns in json gives a way to uniquely encode the columns
// a map is then used (with the encoded key) to store the tuples
func (ts *tupleSet) add(tuple tuple) {
	if len(tuple) != ts.numberOfColumns {
		fatal(ts.collectionName, "expected", ts.numberOfColumns, "columns for", tuple)
	}

	key, err := json.Marshal(tuple[:ts.keyEnds])
	if err != nil {
		panic(err)
	}

	ts.tuples[string(key)] = tuple
}

func (ts *tupleSet) message() messageContainer {
	message := messageContainer{
		operation:  "data",
		collection: ts.collectionName,
		data:       []tuple{},
	}

	for _, tuple := range ts.tuples {
		message.data = append(message.data, tuple)
	}

	return message
}

func collectionHandler(collectionName string, s *service.Service, channels channels) {
	controlinfo(collectionName, "started")
	input := channels.collections[collectionName]
	c := s.Collections[collectionName]

	immediates := ruleChannels(false, collectionName, s, channels)
	deferreds := ruleChannels(true, collectionName, s, channels)

	controlinfo(collectionName, "sends to", immediates, deferreds)

	data := tupleSet{
		tuples:          map[string]tuple{},
		keyEnds:         len(c.Key),
		numberOfColumns: len(c.Key) + len(c.Data),
		collectionName:  collectionName,
	}

	for {
		message := <-input
		controlinfo(collectionName, "received", message)

		switch message.operation {
		case "immediate":
			// info(collectionName, "immediate")
			sendToAll(data.message(), immediates)
			channels.finished <- true
			controlinfo(collectionName, "finished with", message)
		case "deferred":
			// info(collectionName, "deferred")
			controlinfo(collectionName, "sending to", deferreds)
			sendToAll(data.message(), deferreds)
			switch c.Type {
			case service.CollectionInput, service.CollectionOutput, service.CollectionScratch, service.CollectionChannel:
				// temporary collections are emptied
				data.tuples = map[string]tuple{}
			case service.CollectionTable:
				// persistent collections
				// no-op
			default:
				fatal(collectionName, "unhandled collection type", c.Type)
			}
			channels.finished <- true
			controlinfo(collectionName, "finished with", message)
		case "data", "<~":
			flowinfo(collectionName, "received", message)
			for _, tuple := range message.data {
				data.add(tuple)
			}
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
