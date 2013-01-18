package executor

import (
	"bytes"
	"fmt"
	"github.com/msgpack/msgpack-go"
	"net"
	"reflect"
	"service"
)

type bud struct {
	address  string
	listener net.PacketConn
	s        *service.Service
	channels channels
}

func (bud *bud) listen() {
	var err error
	bud.listener, err = net.ListenPacket("udp", bud.address)
	if err != nil {
		fatal("budCommunicator", err)
	}
}

func (bud *bud) close() {
	// bud.listener.Close()
}

func (bud *bud) networkReader() {
	buffer := make([]byte, 1024)
	for {
		// receive from network
		n, _, err := bud.listener.ReadFrom(buffer)
		if err != nil {
			panic(err)
		}

		// unmarshal data
		collectionName, tuples := bud.unmarshal(buffer[:n])

		// validate data (only lengths)
		collection, ok := bud.s.Collections[collectionName]
		if !ok {
			// unknown collection, drop message
			continue
		}
		expectedLength := len(collection.Key) + len(collection.Data)
		for _, tuple := range tuples {
			if len(tuple) != expectedLength {
				// don't try to send to channels when data lengths don't match
				networkerror("budInputReader", "expected lenght of", expectedLength, "for", tuple, "for collection", collectionName)
				continue
			}
		}

		// send to correct collection
		channel := bud.channels.collections[collectionName]
		channel <- messageContainer{
			operation:  "<~",
			collection: collectionName,
			data:       tuples,
		}
	}
}

func (bud *bud) unmarshal(buf []byte) (collectionName string, tuples []tuple) {
	// msgpack format from bud:
	// []interface{}{[]byte, []interface{}{COLUMNS}, []interface{}{}}
	// COLUMNS depends on the column types
	// 0: collection name
	// 1: data
	// 2: ??
	msgReflected, _, err := msgpack.Unpack(bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}
	var msg []interface{}
	switch msgTyped := msgReflected.Interface().(type) {
	case []interface{}:
		msg = []interface{}(msgTyped)
	case []uint8:
		// some other message??
		info("budInputReader", "[]uint8: ", string(msgTyped))
		return "", []tuple{}
	default:
		panic(fmt.Sprintf("%v\n", msgReflected))
	}

	collectionName = string(msg[0].([]byte))

	tuples = []tuple{}
	switch r := msg[1].(type) {
	case []interface{}:
	SingleInterfaceLoop:
		for _, row := range r {
			switch rowTyped := row.(type) {
			case []interface{}:
				tuples = append(tuples, rowTyped)
			case []uint8: // really a single row
				rowFilled := []interface{}{}
				for _, column := range r {
					rowFilled = append(rowFilled, column)
				}
				tuples = append(tuples, rowFilled)
				break SingleInterfaceLoop
			default:
				panic("unknown type:" + reflect.TypeOf(row).String())
			}
		}
	default:
		panic("unknown type:" + reflect.TypeOf(r).String())
	}

	return collectionName, tuples
}

func (bud *bud) marshal(collectionName string, tuple tuple) []byte {
	// create the payload
	outputMessage := bytes.NewBuffer([]byte{})

	// msgpack format from bud:
	// []interface{}{[]byte, []interface{}{COLUMNS}, []interface{}{}}
	// COLUMNS depends on the column types
	// 0: collection name
	// 1: data
	// 2: ??

	// collection name []byte
	// collectionName := []byte(collectionName)

	// // data [][]interface{}
	// data := message.data

	// // ?? []interface{}
	// part3 := []interface{}{}

	_, err := msgpack.Pack(outputMessage, []interface{}{
		collectionName,
		tuple,
		[]interface{}{},
	})
	if err != nil {
		panic(err)
	}

	return outputMessage.Bytes()
}

func (bud *bud) send(message messageContainer) {
	flowinfo("budCommunicator", "sending", message.String())
	// make sure the collection is known
	collection, ok := bud.s.Collections[message.collection]
	if !ok {
		panic(fmt.Sprintf("unknown collection from %v", message))
	}

	// find the address column
	addressColumn := -1
	for index, name := range collection.Key {
		if name[0] == '@' {
			addressColumn = index
			break
		}
	}
	if addressColumn == -1 {
		panic("no address column for collection " + message.collection)
	}

	for _, tuple := range message.data {
		// get the address to send to
		address, err := net.ResolveUDPAddr("udp", string(tuple[addressColumn].([]uint8)))
		if err != nil {
			panic(err)
		}

		// don't send to self
		if address.String() == bud.address {
			continue
		}

		// marshal the tuple
		marshalled := bud.marshal(message.collection, tuple)

		// send the tuple
		_, err = bud.listener.WriteTo(marshalled, address)
		if err != nil {
			panic(err)
		}

		flowinfo("budCommunicator", "sent", tuple.String(), "to", string(tuple[addressColumn].([]byte)))
	}
}

func budCommunicator(s *service.Service, channels channels, address string) {
	bud := bud{
		address:  address,
		s:        s,
		channels: channels,
	}

	bud.listen()
	defer bud.close()

	go bud.networkReader()

	controlinfo("budCommunicator", "started")
	for {
		message := <-channels.distribution
		controlinfo("budCommunicator", "received", message)

		switch message.operation {
		case "immediate", "deferred":
			channels.control <- messageContainer{operation: "done", collection: "budCommunicator"}
			controlinfo("budCommunicator", "finished with", message)
		case "data":
			flowinfo("budCommunicator", "received", message)
			bud.send(message)
		default:
			fatal("budCommunicator", "unhandled message:", message)
		}
	}
}
