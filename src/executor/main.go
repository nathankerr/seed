package main

import (
	"bytes"
	"fmt"
	"github.com/msgpack/msgpack-go"
	"io/ioutil"
	// "log"
	"net"
	net_transform "network"
	// "os"
	"service"
	"time"
)

type input struct {
	channel string
	data    map[string]string
}

type message struct {
	operation  string // "<-", "<+", ... , "", "step", "done"
	collection collection
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

	// things for handling the main loop
	main := make(chan message)

	// make the channels to connect to/from the channels
	collection_inputs := make(map[string]chan message)
	for name, _ := range kvs.Collections {
		channel := make(chan message)
		collection_inputs[name] = channel
	}

	// launch the input handler
	bud_input := make(chan message)
	// handler_inputs = append(handler_inputs, bud_input)
	input_collections := make(map[string]chan<- message)
	for name, channel := range collection_inputs {
		if kvs.Collections[name].Type == service.CollectionChannel {
			input_collections[name] = channel
		}
	}
	go budInputHandler("127.0.0.1:3000", bud_input, input_collections, main)

	// Process the rules
	// 1. make input channels for the rules
	// 2. collect the rule input channels for each of the collections which feed data to the channels
	// 3. launch the rule handlers
	rule_inputs := make([]chan message, len(kvs.Rules))
	collection_outputs := make(map[string][]chan<- message)
	for rule_num, rule := range kvs.Rules {
		// make input channel
		channel := make(chan message)
		rule_inputs[rule_num] = channel

		// add the rule_input to the set of channels for the collection handlers
		for _, requires := range rule.Requires() {
			collection_outputs[requires] = append(collection_outputs[requires], rule_inputs[rule_num])
		}

		// launch handler
		go ruleHandler(rule_num,
			rule_inputs[rule_num],
			collection_inputs[rule.Supplies],
			kvs,
			rule,
			main)
	}

	// launch the collection handlers
	for name, collection := range kvs.Collections {
		go collectionHandler(name,
			collection_inputs[name],
			collection_outputs[name],
			collection,
			main)
	}

	// the timestep loop
	step := message{operation: "step"}
	for {
		startTime := time.Now()
		time.Sleep(2 * time.Second)
		info("main", "------------------------------------")

		// tell all handlers to step
		for collection_name, collection_channel := range collection_inputs {
			select {
			case collection_channel <- step:
			}
			info("main", "sent step to", collection_name)
		}

		// wait for all the handlers to finish the step
		finished := 0
		for finished != len(collection_inputs)+len(rule_inputs) {
			select {
			case <-main:
				finished++
			}
			info("main", finished, "of", len(collection_inputs)+len(rule_inputs))
		}

		info("main", "step took", time.Since(startTime))
	}
}

// try out a bud-compatible network interface
func budInputHandler(addr string, input <-chan message, collections map[string]chan<- message, main chan<- message) {
	// listen on the network
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

		info("budInput", "sent", to_send, "to", collection_name)
	}
}

func collectionHandler(name string, input <-chan message, outputs []chan<- message, collection *service.Collection, main chan<- message) {
	c := newCollection(name, collection)

	// timestep loop
	// sends to outputs when "step" is received
	// otherwise, handles operation on the collection
	for {
		info(name, "ready")
		incoming_message := <-input
		switch incoming_message.operation {
		case "step":
			info(name, "step")
			outgoing_message := message{
				operation:  "",
				collection: *c,
			}
			for _, output := range outputs {
				output <- outgoing_message
			}
			main <- message{operation: "done"}
			info(name, "done")
		case "<+", "<~":
			info(name, "merge", incoming_message.collection.rows)
			c.merge(&incoming_message.collection)
			info(name, "has", c.rows)
		case "<-":
			info(name, "delete", incoming_message.collection.rows)
			c.delete(&incoming_message.collection)
		default:
			info(name, "unhandled operation:", incoming_message.operation)
		}
	}
}

func ruleHandler(rule_num int, input <-chan message, output chan<- message, service *service.Service, rule *service.Rule, main chan<- message) {
	// timestep loop
	// 1. receive required collections
	// 2. run rule
	// 3. send resulting collection
	// 4. send finished timestep to main
	for {
		info(rule_num, "ready")
		// create collections
		collections := make(map[string]*collection)

		// wait for all the required collections to arrive
		for len(collections) != len(rule.Requires()) {
			incoming_message := <-input
			collections[incoming_message.collection.name] = &incoming_message.collection
			info(rule_num, len(collections), "of", len(rule.Requires()))
		}

		// run rule
		result := runRule(collections, service, rule)

		// spend result
		output <- message{
			operation:  rule.Operation,
			collection: *result,
		}
		info(rule_num, "sent")

		// tell main step is finished
		main <- message{
			operation: "done",
		}
		info(rule_num, "done")
	}
}
