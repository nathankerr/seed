package executor

import (
	"encoding/json"
	service "github.com/nathankerr/seed"
)

type tupleSet struct {
	tuples          map[string]service.Tuple
	keyEnds         int
	numberOfColumns int
	collectionName  string
}

// tuples are unique according to their key columns
// the key columns are a subset of the columns starting at the beginning
// encoding the key columns in json gives a way to uniquely encode the columns
// a map is then used (with the encoded key) to store the tuples
func (ts *tupleSet) add(tuple service.Tuple) {
	if len(tuple) != ts.numberOfColumns {
		fatal(ts.collectionName, "expected", ts.numberOfColumns, "columns for", tuple)
	}

	key, err := json.Marshal(tuple[:ts.keyEnds])
	if err != nil {
		panic(err)
	}

	ts.tuples[string(key)] = tuple
}

func (ts *tupleSet) message() MessageContainer {
	message := MessageContainer{
		Operation:  "data",
		Collection: ts.collectionName,
		Data:       []service.Tuple{},
	}

	for _, tuple := range ts.tuples {
		message.Data = append(message.Data, tuple)
	}

	return message
}

func collectionHandler(collectionName string, s *service.Seed, channels Channels) {
	controlinfo(collectionName, "started")
	input := channels.Collections[collectionName]
	c := s.Collections[collectionName]

	immediates := ruleChannels(false, collectionName, s, channels)
	deferreds := ruleChannels(true, collectionName, s, channels)
	if c.Type == service.CollectionChannel {
		deferreds = append(deferreds, channels.Distribution)
	}

	controlinfo(collectionName, "sends to", immediates, deferreds)

	data := tupleSet{
		tuples:          map[string]service.Tuple{},
		keyEnds:         len(c.Key),
		numberOfColumns: len(c.Key) + len(c.Data),
		collectionName:  collectionName,
	}

	for {
		message := <-input
		controlinfo(collectionName, "received", message)

		switch message.Operation {
		case "immediate":
			// info(collectionName, "immediate")
			dataMessage := data.message()
			sendToAll(dataMessage, immediates)
			dataMessage.Operation = "done"
			channels.Control <- dataMessage
			controlinfo(collectionName, "finished with", message)
		case "deferred":
			// info(collectionName, "deferred")
			controlinfo(collectionName, "sending to", deferreds)
			dataMessage := data.message()
			sendToAll(dataMessage, deferreds)
			switch c.Type {
			case service.CollectionInput, service.CollectionOutput, service.CollectionScratch, service.CollectionChannel:
				// temporary collections are emptied
				data.tuples = map[string]service.Tuple{}
			case service.CollectionTable:
				// persistent collections
				// no-op
			default:
				fatal(collectionName, "unhandled collection type", c.Type)
			}
			dataMessage.Operation = "done"
			channels.Control <- dataMessage
			controlinfo(collectionName, "finished with", message)
		case "data", "<~":
			flowinfo(collectionName, "received", message.String())
			for _, tuple := range message.Data {
				data.add(tuple)
			}
		default:
			fatal(collectionName, "unhandled message:", message)
		}

	}
}

func ruleChannels(deferred bool, collectionName string, s *service.Seed, channels Channels) []chan<- MessageContainer {
	ruleChannels := []chan<- MessageContainer{}
	for ruleNum, rule := range s.Rules {
		if (deferred && rule.Operation == "<=") || (!deferred && rule.Operation != "<=") {
			continue
		}
		for _, collection := range rule.Requires() {
			if collection == collectionName {
				ruleChannels = append(ruleChannels, channels.Rules[ruleNum])
				break
			}
		}
	}
	return ruleChannels
}
