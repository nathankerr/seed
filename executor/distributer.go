package executor

// TODO:
// - add error logging, etc to register message handling

import (
	"github.com/nathankerr/seed"
	"strings"
)

type communicator struct {
	prefix  string
	channel chan<- MessageContainer
}

func distributer(s *seed.Seed, channels Channels) {
	controlinfo("_distributer", "started")
	traces := []chan<- MessageContainer{}
	controls := []chan<- MessageContainer{}
	communicators := []communicator{}
	data := map[string][]seed.Tuple{}

	for {
		select {
		case message := <-channels.Distribution: // from the collections and rules
			flowinfo("_distributer", "received", message)
			switch message.Operation {
			case "immediate":
				for collectionName, _ := range s.Collections {
					channel := channels.Collections[collectionName]
					message := MessageContainer{
						Operation:  "data",
						Collection: collectionName,
						Data:       data[collectionName],
					}
					channel <- message
					flowinfo("_distributer", "sent", message, "to", collectionName)
					data[collectionName] = []seed.Tuple{}
				}
				channels.Control <- MessageContainer{Operation: "done", Collection: "_distributer"}
			case "deferred":
				channels.Control <- MessageContainer{Operation: "done", Collection: "_distributer"}
			case "data":
				// TODO: send message to correct communicator
				collection, ok := s.Collections[message.Collection]
				if !ok {
					continue
				}

				if collection.Type == seed.CollectionChannel {
					addressColumn, ok := collection.AddressColumn()
					if !ok {
						continue
					}

					for _, tuple := range message.Data {
						for _, communicator := range communicators {
							address, ok := tuple[addressColumn].(string)
							if !ok {
								continue
							}
							if strings.HasPrefix(address, communicator.prefix) {
								communicator.channel <- MessageContainer{
									Operation:  "data",
									Collection: message.Collection,
									Data:       []seed.Tuple{tuple},
								}
							}
						}
					}
				}
			case "register":
				switch message.Collection {
				case "_trace":
					// Data holds a channel to be added to the trace list
					channel, ok := extractChannel(message)
					if !ok {
						continue
					}
					traces = append(traces, channel)
				case "_control":
					// Data hold a channel to be added to the control list
					channel, ok := extractChannel(message)
					if !ok {
						continue
					}
					controls = append(controls, channel)
				default:
					// if strings.HasPrefix(addressfield, message.Collection) then send collection data to the channel
					// Data holds a channel
					channel, ok := extractChannel(message)
					if !ok {
						continue
					}
					communicators = append(communicators, communicator{message.Collection, channel})
				}
			case "control":
				// Collection holds the control operation to send
			case "insert":
				// Collection holds the collection, Data the data
				// data is inserted at the next "immediate"
				if _, ok := data[message.Collection]; !ok {
					continue
				}

				data[message.Collection] = append(data[message.Collection], message.Data...)
			default:
				fatal("distributer", "unhandled message:", message)
			}
		}
	}
}

func extractChannel(message MessageContainer) (chan<- MessageContainer, bool) {
	if len(message.Data) != 1 {
		return nil, false
	}

	tuple := message.Data[0]

	if len(tuple) != 1 {
		return nil, false
	}

	channel, ok := tuple[0].(chan<- MessageContainer)
	return channel, ok
}
