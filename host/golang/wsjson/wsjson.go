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
	"encoding/json"
	"fmt"
	"github.com/nathankerr/seed"
	executor "github.com/nathankerr/seed/host/golang"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
)

var registerSocket = make(chan socket)
var incomingMessage = make(chan executor.MessageContainer)

type socket struct {
	io.ReadWriter
	address      string
	done         chan bool
	localaddress string
}

func Communicator(s *seed.Seed, channels executor.Channels, address string) {
	info("starting wsjson server")
	go server(address)
	info("server started")

	fromDistribution := make(chan executor.MessageContainer)
	channels.Distribution <- executor.MessageContainer{
		Operation:  "register",
		Collection: "",
		Data:       []seed.Tuple{seed.Tuple{fromDistribution}},
	}

	sockets := map[string]socket{}      // address: socket
	localAddresses := map[string]bool{} // list of addresses this communicator responds to, messages to these addresses will be dropped
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
		case message := <-fromDistribution:
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

					thisSocket, ok := sockets[tupleAddress]
					if !ok {
						// skip addresses without sockets
						// this also drops tuples sent to the local address
						if localAddresses[tupleAddress] {
							continue
						}

						log("socket for address \""+tupleAddress+"\" not found", thisSocket)

						ws, err := websocket.Dial(fmt.Sprintf("ws://%s/wsjson", tupleAddress), "", "http://locahost:3000")
						if err != nil {
							log(err)
						}

						thisSocket = socket{
							ws,
							tupleAddress,
							nil,
							"",
						}

						thisSocket.Write([]byte("\"" + tupleAddress + "\""))

						sockets[tupleAddress] = thisSocket

						go incomingMessageReader(ws)
						// continue
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

					_, err = thisSocket.Write(marshalled)
					if err != nil {
						// close the handler
						thisSocket.done <- true

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
			localAddresses[socket.localaddress] = true
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
		log(err, string(raw[:n]))
	}

	registerSocket <- socket{ws, address, done, ws.LocalAddr().String()}

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
