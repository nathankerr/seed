package executor

import (
	"bytes"
	"fmt"
	"github.com/msgpack/msgpack-go"
	"net"
	"reflect"
	"service"
)

// try out a bud-compatible network interface
func budInputHandler(listener net.PacketConn, input <-chan message, collections map[string]chan<- message, main chan<- message, seed *service.Service) {
	buf := make([]byte, 1024)
	for {
		n, _, err := listener.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		// fmt.Printf("received %#v\n", buf[:n])

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
		// info("budInput", collection_name)

		collection_channel, ok := collections[collection_name]
		if !ok {
			fmt.Println("no channel for", collection_name)
		}

		// create the collection to send
		// row := msg[1].([]interface{})
		// key := make([]string, len(row))
		// for i, _ := range row {
		// 	key[i] = fmt.Sprintf("column%d", i)
		// }
		// to_send := newCollectionFromRaw(
		// 	collection_name,
		// 	key,
		// 	nil,
		// 	[][]interface{}{row},
		// )
		to_send := newCollection(collection_name, seed.Collections[collection_name])
		switch r := msg[1].(type) {
		// case [][]interface{}:
		// 	fmt.Println("double array")
		// 	to_send.addRows(r)
		case []interface{}:
			// fmt.Println("have", r)
		SingleInterfaceLoop:
			for _, row := range r {
				switch row_typed := row.(type) {
				case []interface{}:
					to_send.addRow(row_typed)
				case []uint8: // really a single row
					row_filled := []interface{}{}
					for _, column := range r {
						row_filled = append(row_filled, column)
					}
					to_send.addRow(row_filled)
					break SingleInterfaceLoop
				default:
					panic("unknown type:" + reflect.TypeOf(row).String())
				}
			}
		default:
			panic("unknown type:" + reflect.TypeOf(r).String())
		}

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
	collections := make(map[string]*collection)
	for {
		message := <-input
		switch message.operation {
		case "": // is blank because all channels send to the output handlers; collections don't know what their operations are
			collection, ok := collections[message.collection.name]
			if !ok {
				collections[message.collection.name] = &message.collection
			} else {
				collection.merge(&message.collection)
			}
		case "step": // when stepping, send all collected rows then empty set to send
			// send rows
			for _, collection := range collections {
				monitorinfo("budOutput", collection.name, ": ", collection.String())

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

					// don't send to self
					if address.String() == "127.0.0.1:3000" {
						continue
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
					collection_name := []byte(collection.name)

					// data [][]interface{}
					data := []interface{}{}

					for _, column := range row {
						data = append(data, column)
					}

					// ?? []interface{}
					part3 := []interface{}{}

					_, err = msgpack.Pack(outputMessage, []interface{}{collection_name, data, part3})
					if err != nil {
						panic(err)
					}

					// send the message
					_, err = conn.WriteTo(outputMessage.Bytes(), address)
					if err != nil {
						panic(err)
					}

					info("budOutput", "sent to", string(row[addressColumn].([]byte)), row)
				}
			}
			collections = make(map[string]*collection)
		default:
			panic("unhandled operation: " + message.operation)
		}
	}
}
