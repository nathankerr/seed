package main

import (
	"bytes"
	"fmt"
	"github.com/msgpack/msgpack-go"
	"io/ioutil"
	"net"
	"service"
	"time"
)

type input struct {
	channel string
	data    map[string]string
}

type message struct {
	collection string
	data       map[string]string
}

func main() {
	// load the service
	kvs_source, err := ioutil.ReadFile("kvs.seed")
	if err != nil {
		panic(err)
	}
	kvs := service.Parse("kvs.seed", string(kvs_source))

	for num, rule := range kvs.Rules {
		fmt.Printf("%d\t%s\n", num, rule)
	}
	fmt.Println("-------------------------------")

	// make the channels to connect to/from the channels
	collection_inputs := make(map[string]chan message)
	for name, _ := range kvs.Collections {
		collection_inputs[name] = make(chan message)
	}

	rule_inputs := make([]chan message, len(kvs.Rules))
	collection_outputs := make(map[string][]chan<- message)
	for rule_num, rule := range kvs.Rules {
		// make input channel
		rule_inputs[rule_num] = make(chan message)

		for _, requires := range rule.Requires() {
			collection_outputs[requires] = append(collection_outputs[requires], rule_inputs[rule_num])
		}

		// launch handler
		go ruleHandler(rule_num,
			rule_inputs[rule_num],
			collection_inputs[rule.Supplies],
			rule)
	}

	for name, collection := range kvs.Collections {
		go collectionHandler(name,
			collection_inputs[name],
			collection_outputs[name],
			collection)
	}

	// example input
	input_message := input{channel: "kvget", data: map[string]string{"key": "123"}}
	collection_inputs["kvget"] <- message{data: input_message.data}
	collection_inputs["kvget"] <- message{data: map[string]string{"key": "456"}}

	// listen_addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:3000")
	// if err != nil {
	// 	panic(err)
	// }

	// listener, err := net.ListenUDP("udp", listen_addr)
	// if err != nil {
	// 	panic(err)
	// }

	listener, err := net.ListenPacket("udp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	buf := make([]byte, 512)
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
	msg := msg_reflected.Interface().([]interface{})
	channel_name := string(msg[0].([]byte))
	fmt.Println(channel_name)
	row := msg[1].([]interface{})
	fmt.Println(row)

	// // [<[]uint8 Value> <[]reflect.Value Value> <[]reflect.Value Value>]
	// fmt.Printf("%#v\n", msg.Interface())
	// // []interface {}{[]byte{0x73, 0x65, 0x74}, []interface {}{[]byte{0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x33, 0x30, 0x30, 0x30}, 123, []byte{0x6b, 0x65, 0x79}, 12}, []interface {}{}}
	// msg1 := msg.Interface().([]interface{})
	// fmt.Printf("%d\n", len(msg1))
	// // 3
	// msg1_0 := msg1[0].([]byte)
	// fmt.Printf("%#v\n", string(msg1_0))
	// // "set"
	// msg1_1 := msg1[1].([]interface{})
	// fmt.Printf("%#v\n", msg1_1)
	// //[]interface {}{[]byte{0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, 0x3a, 0x33, 0x30, 0x30, 0x30}, 123, []byte{0x6b, 0x65, 0x79}, 12}
	// msg1_1_0 := msg1_1[0].([]byte)
	// fmt.Printf("%#v\n", string(msg1_1_0))

	// wait for the example input to be processed before exiting
	time.Sleep(time.Second)
	fmt.Println("done")

}

func collectionHandler(name string, input <-chan message, outputs []chan<- message, collection *service.Collection) {
	for {
		select {
		case message := <-input:
			fmt.Printf("[%s] %s\n", name, message.data)
			for _, output := range outputs {
				output <- message
			}
		}
	}
}

func ruleHandler(rule_num int, input <-chan message, output chan<- message, rule *service.Rule) {
	for {
		select {
		case message := <-input:
			fmt.Printf("[%d] %s\n", rule_num, message.data)
			output <- message
		}
	}
}
