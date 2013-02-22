package wsjson

// note:
// The javascript websocket api does not provide a method
// to determine the source port of the socket connection.
// Possible work-arounds:
// - server returns the "remote address" when a connection
//   is established
// - drop the requirement for ip addresses and use names
//   a random (hopefully unique) name can be generated with:
//   Math.random().toString(36).substr(2,5)
//   which generates a 5 char long base 36 number, which is
//   represented using numbers and letters

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/nathankerr/seed"
	"github.com/nathankerr/seed/executor"
	"io"
	"net/http"
)

var registerSocket = make(chan socket)
var incomingMessage = make(chan executor.MessageContainer)

type socket struct {
	io.ReadWriter
	address string
	done    chan bool
}

func Communicator(s *seed.Seed, channels executor.Channels, address string) {
	info("starting wsjson server")
	go server(address)
	info("server started")

	sockets := map[string]socket{} // address: socket
	for {
		info("Communicator")
		select {
		case message := <-incomingMessage:
			info("incoming", message)
			channel, ok := channels.Collections[message.Collection]
			if !ok {
				continue
			}
			channel <- message
		case message := <-channels.Distribution:
			info("distribution")
			switch message.Operation {
			case "immediate", "deferred":
				channels.Control <- executor.MessageContainer{Operation: "done", Collection: "wsjsonCommunicator"}
			case "data":
				// ignore messages with out a data payload
				if len(message.Data) == 0 {
					continue
				}
				info("data", message)

				// send the message to the correct socket (by address)
				// find the address column
				collection := s.Collections[message.Collection]
				addressColumn := -1
				for index, name := range collection.Key {
					if name[0] == '@' {
						addressColumn = index
						break
					}
				}
				if addressColumn == -1 {
					panic("no address column for collection " + message.Collection)
				}

				// split the incoming message into messages for each address
				// info("split")
				// messages := map[string]*executor.MessageContainer{}
				// for _, tuple := range message.Data {
				// 	tupleAddress := tuple[addressColumn].(string)
				// 	if tupleAddress == address {
				// 		continue
				// 	}

				// 	new_message, ok := messages[tupleAddress]
				// 	if !ok {
				// 		messages[tupleAddress] = &executor.MessageContainer{
				// 			Operation: message.Operation,
				// 			Collection: message.Collection,
				// 			Data: []seed.Tuple{},
				// 		}
				// 		new_message = messages[tupleAddress]
				// 	}

				// 	new_message.Data = append(new_message.Data, tuple)
				// }

				// send the messages
				for _, tuple := range message.Data {
					info("Sending: ", message.String())
					tupleAddress := tuple[addressColumn].(string)
					if tupleAddress == address {
						// log("skipping", tuple)
						continue
					}

					socket, ok := sockets[tupleAddress]
					if !ok {
						// skip addresses without sockets
						log("socket for address \""+tupleAddress+"\" not found", socket)
						continue
					}

					toSend := executor.MessageContainer{
						Operation:  message.Operation,
						Collection: message.Collection,
						Data:       []seed.Tuple{tuple},
					}

					marshalled, err := json.Marshal(toSend)
					if err != nil {
						panic(err)
					}

					_, err = socket.Write(marshalled)
					if err != nil {
						// close the handler
						socket.done <- true

						// remove from the list of sockets
						delete(sockets, tupleAddress)
					}
				}
				info("wsjson", "data processed")
			default:
				panic(message.Operation)
			}
		case socket := <-registerSocket:
			// add socket to the list of known sockets
			info("register", socket.address)
			sockets[socket.address] = socket
		}
	}
}

func server(address string) {
	http.Handle("/wsjson", websocket.Handler(wsHandler))
	err := http.ListenAndServe(address, nil)
	if err != nil {
		fatal("_wsjsonCommunicator", err)
	}
}

func wsHandler(ws *websocket.Conn) {
	done := make(chan bool)

	raw := make([]byte, 1024)
	n, err := ws.Read(raw)
	if err != nil {
		log(err)
		return
	}

	var address string
	err = json.Unmarshal(raw[:n], &address)
	if err != nil {
		log(err)
	}

	registerSocket <- socket{ws, address, done}

	go incomingMessageReader(ws)

	// when this function finishes the socket will be closed
	<-done
	ws.Close()
}

func incomingMessageReader(ws *websocket.Conn) {
	for {
		info("reader")
		raw := make([]byte, 1024)
		n, err := ws.Read(raw)
		if err != nil {
			info(err)
			continue
		}
		info("received:", string(raw[:n]))

		message := executor.MessageContainer{}
		err = json.Unmarshal(raw[:n], &message)
		if err != nil {
			info(err)
		}

		incomingMessage <- message
	}
}
