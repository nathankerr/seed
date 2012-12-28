package main

import (
	"bytes"
	"fmt"
	"github.com/msgpack/msgpack-go"
	"io/ioutil"
	"net"
	net_transform "network"
	"os"
	"service"
	"time"
)

type input struct {
	channel string
	data    map[string]string
}

type message struct {
	collection string
	data       []interface{}
}

func main() {
	// load the service and add a network interface
	kvs_source, err := ioutil.ReadFile("kvs.seed")
	if err != nil {
		panic(err)
	}
	seeds := make(map[string]*service.Service)
	seeds["kvs"] = service.Parse("kvs.seed", string(kvs_source))
	transformed := make(map[string]*service.Service)
	transformed = net_transform.Add_network_interface("kvs", seeds["kvs"], transformed)
	kvs := transformed["KvsServer"]

	// print the rules, so the numbers are known for tracing
	for num, rule := range kvs.Rules {
		fmt.Printf("%d\t%s\n", num, rule)
	}
	fmt.Println("-------------------------------")

	// make the channels to connect to/from the channels
	collection_inputs := make(map[string]chan message)
	for name, _ := range kvs.Collections {
		collection_inputs[name] = make(chan message)
	}

	// Process the rules
	// 1. make input channels for the rules
	// 2. collect the rule input channels for each of the collections which feed data to the channels
	// 3. launch the rule handlers
	rule_inputs := make([]chan message, len(kvs.Rules))
	collection_outputs := make(map[string][]chan<- message)
	for rule_num, rule := range kvs.Rules {
		// make input channel
		rule_inputs[rule_num] = make(chan message)

		// add the 
		for _, requires := range rule.Requires() {
			collection_outputs[requires] = append(collection_outputs[requires], rule_inputs[rule_num])
		}

		// launch handler
		go ruleHandler(rule_num,
			rule_inputs[rule_num],
			collection_inputs[rule.Supplies],
			rule)
	}

	// launch the collection handlers
	for name, collection := range kvs.Collections {
		go collectionHandler(name,
			collection_inputs[name],
			collection_outputs[name],
			collection)
	}

	bud_input := make(chan message)
	input_collections := make(map[string]chan<- message)
	for name, channel := range collection_inputs {
		if kvs.Collections[name].Type == service.CollectionChannel {
			input_collections[name] = channel
		}
	}
	budInterface("127.0.0.1:3000", bud_input, input_collections)
}

func budInterface(addr string, input <-chan message, collections map[string]chan<- message) {
	// try out a bud-compatible network interface
	listener, err := net.ListenPacket("udp", "localhost:3000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

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
		msg := msg_reflected.Interface().([]interface{})
		collection_name := string(msg[0].([]byte))

		fmt.Println("received for", collection_name)
		collection_channel, ok := collections[collection_name]
		if !ok {
			fmt.Println("no channel for", collection_name)
		}

		collection_channel <- message{
			collection: collection_name,
			data: msg[1].([]interface{}),
		}
		fmt.Println("sent")
		time.Sleep(10*time.Second)
		os.Exit(0)
	}
}

func collectionHandler(name string, input <-chan message, outputs []chan<- message, collection *service.Collection) {
	for {
		select {
		case message := <-input:
			fmt.Printf("[%s] %v\n", name, message.data)
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
			fmt.Printf("[%d] %v\n", rule_num, message.data)
			output <- message
		}
	}
}
