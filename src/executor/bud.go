package main

import (
	"bytes"
	"fmt"
	"github.com/msgpack/msgpack-go"
	"net"
)

// try out a bud-compatible network interface
func budInputHandler(listener net.PacketConn, input <-chan message, collections map[string]chan<- message, main chan<- message) {
	buf := make([]byte, 1024)
	for {
		n, _, err := listener.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		// msgpack format from bud:
		// []interface{}{[]byte, []interface{}{COLUMNS}, []interface{}{}}
		// COLUMNS depends on the column types
		// 0: collection name
		// 1: data
		// 2: ??
		msg_reflected, _, err := msgpack.Unpack(bytes.NewBuffer(buf[:n]))
		if err != nil {
			panic(err)
		}
		var msg []interface{}
		switch msg_typed := msg_reflected.Interface().(type) {
		case []interface{}:
			msg = []interface{}(msg_typed)
		case []uint8:
			// some other message??
			info("budInput", "[]uint8: ", string(msg_typed))
			continue
		default:
			panic(fmt.Sprintf("%v\n", msg_reflected))
		}
		// msg := msg_reflected.Interface().([]interface{})
		collection_name := string(msg[0].([]byte))
		info("budInput", collection_name)

		collection_channel, ok := collections[collection_name]
		if !ok {
			fmt.Println("no channel for", collection_name)
		}

		// create the collection to send
		row := msg[1].([]interface{})
		key := make([]string, len(row))
		for i, _ := range row {
			key[i] = fmt.Sprintf("column%d", i)
		}
		to_send := newCollectionFromRaw(
			"bud input",
			key,
			nil,
			[][]interface{}{row},
		)

		// send the collection
		collection_channel <- message{
			operation:  "<+",
			collection: *to_send,
		}

		flowinfo("budInput", "sent", to_send, "to", collection_name)
	}
}

// a bud compatible sender network interface
func budOutputHandler(conn net.PacketConn, input <-chan message) {
	for {
		message := <-input

		monitorinfo("budOutput", "has", message)
		collection := message.collection

		// find the address column
		addressColumn := -1
		for name, index := range collection.columns {
			if name[0] == '@' {
				addressColumn = index
				break
			}
		}
		if addressColumn == -1 {
			panic("no address column in " + collection.name)
		}

		for _, row := range collection.rows {
			// get the address to send to
			address, err := net.ResolveUDPAddr("udp", string(row[addressColumn].([]uint8)))
			if err != nil {
				panic(err)
			}

			// create the payload
			outputMessage := bytes.NewBuffer([]byte{})

			// msgpack format from bud:
			// []interface{}{[]byte, []interface{}{COLUMNS}, []interface{}{}}
			// COLUMNS depends on the column types
			// 0: collection name
			// 1: data
			// 2: ??

			// collection name []byte
			_, err = msgpack.PackBytes(outputMessage, []byte(collection.name))
			if err != nil {
				panic(err)
			}

			// data [][]interface{}
			data := [][]interface{}{}
			for _, row := range collection.rows {
				data = append(data, row)
			}
			_, err = msgpack.Pack(outputMessage, data)
			if err != nil {
				panic(err)
			}

			// ?? []interface{}
			_, err = msgpack.Pack(outputMessage, []interface{}{})
			if err != nil {
				panic(err)
			}

			// send the message
			_, err = conn.WriteTo(outputMessage.Bytes(), address)
			if err != nil {
				panic(err)
			}

			info("budOutput", "sent to", string(row[addressColumn].([]byte)))

		}
	}
}