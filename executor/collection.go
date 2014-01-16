package executor

import (
	"encoding/json"
	"github.com/nathankerr/seed"
)

type tupleSet struct {
	tuples          map[string]seed.Tuple
	keyEnds         int
	numberOfColumns int
	collectionName  string
}

// tuples are unique according to their key columns
// the key columns are a subset of the columns starting at the beginning
// encoding the key columns in json gives a way to uniquely encode the columns
// a map is then used (with the encoded key) to store the tuples
func (ts *tupleSet) add(tuple seed.Tuple) {
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
		Data:       []seed.Tuple{},
	}

	for _, tuple := range ts.tuples {
		message.Data = append(message.Data, tuple)
	}

	return message
}

func collectionHandler(collectionName string, s *seed.Seed, channels Channels) {
	controlinfo(collectionName, "started")
	input := channels.Collections[collectionName]
	c := s.Collections[collectionName]

	inputsNeeded := 1 // start with the distributer
	for _, rule := range s.Rules {
		if rule.Supplies == collectionName && rule.Operation == "<=" {
			inputsNeeded++
		}
	}
	inputsReceived := 0

	immediates := ruleChannels(false, collectionName, s, channels)
	deferreds := ruleChannels(true, collectionName, s, channels)

	//controlinfo(collectionName, "sends to", immediates, deferreds)

	data := tupleSet{
		tuples:          map[string]seed.Tuple{},
		keyEnds:         len(c.Key),
		numberOfColumns: len(c.Key) + len(c.Data),
		collectionName:  collectionName,
	}

	for {
		message := <-input
		flowinfo(collectionName, "received", message)

		switch message.Operation {
		case "immediate":
			controlinfo(collectionName, "immediate")

			//get the data that has not come yet
			flowinfo(collectionName, "have ", inputsReceived, "of ", inputsNeeded)
			for inputsReceived < inputsNeeded {
				message := <-input
				switch message.Operation {
				case "data":
					for _, tuple := range message.Data {
						data.add(tuple)
					}
					inputsReceived++
				default:
					fatal(collectionName, "unhandled message", message)
				}
				flowinfo(collectionName, "have ", inputsReceived, "of ", inputsNeeded)
			}
			inputsReceived = 0

			dataMessage := data.message()
			sendToAll(dataMessage, immediates)
			dataMessage.Operation = "done"
			channels.Control <- dataMessage
			controlinfo(collectionName, "finished with", message)
		case "deferred":
			controlinfo(collectionName, "deferred")
			controlinfo(collectionName, "sending to", deferreds)
			inputsReceived = 0
			dataMessage := data.message()
			sendToAll(dataMessage, deferreds)
			switch c.Type {
			case seed.CollectionInput, seed.CollectionOutput, seed.CollectionScratch, seed.CollectionChannel:
				// temporary collections are emptied
				data.tuples = map[string]seed.Tuple{}
			case seed.CollectionTable:
				// persistent collections
				// no-op
			default:
				fatal(collectionName, "unhandled collection type", c.Type)
			}
			dataMessage.Operation = "done"
			channels.Control <- dataMessage
			flowinfo(collectionName, "sent", dataMessage)
			controlinfo(collectionName, "finished with", message)
		case "data", "<~":
			inputsReceived++
			flowinfo(collectionName, "received", inputsReceived, "of", inputsNeeded, ":", message.String())
			for _, tuple := range message.Data {
				data.add(tuple)
			}
		default:
			fatal(collectionName, "unhandled message:", message)
		}

	}
}

func ruleChannels(deferred bool, collectionName string, s *seed.Seed, channels Channels) []chan<- MessageContainer {
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
