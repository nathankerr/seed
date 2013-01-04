package executor

import (
	"fmt"
	"net"
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

func Execute(s *service.Service) {
	// print the rules, so the numbers are known for tracing
	for num, rule := range s.Rules {
		fmt.Printf("%d\t%s\n", num, rule)
	}
	fmt.Println("-------------------------------")

	// things for handling the main loop
	main := make(chan message)

	// make the channels to connect to/from the channels
	collection_inputs := make(map[string]chan message)
	collection_outputs := make(map[string][]chan<- message)
	bud_output := make(chan message)
	for name, collection := range s.Collections {
		channel := make(chan message)
		collection_inputs[name] = channel

		if collection.Type == service.CollectionChannel {
			collection_outputs[name] = append(collection_outputs[name], bud_output)
		}
	}

	// launch the bud input and output handlers
	bud_input := make(chan message)
	input_collections := make(map[string]chan<- message)
	for name, channel := range collection_inputs {
		if s.Collections[name].Type == service.CollectionChannel {
			input_collections[name] = channel
		}
	}
	// the PacketConn for both
	listener, err := net.ListenPacket("udp", "localhost:3000")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	go budInputHandler(listener, bud_input, input_collections, main)
	go budOutputHandler(listener, bud_output)

	// Process the rules
	// 1. make input channels for the rules
	// 2. collect the rule input channels for each of the collections which feed data to the channels
	// 3. launch the rule handlers
	rule_inputs := make([]chan message, len(s.Rules))
	for rule_num, rule := range s.Rules {
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
			s,
			rule,
			main)
	}

	// launch the collection handlers
	for name, collection := range s.Collections {
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
		info("main", "--------------------------------------------")

		// tell all handlers to step
		for collection_name, collection_channel := range collection_inputs {
			select {
			case collection_channel <- step:
			}
			flowinfo("main", "sent step to", collection_name)
		}

		// wait for all the handlers to finish the step
		finished := 0
		for finished != len(collection_inputs)+len(rule_inputs) {
			select {
			case <-main:
				finished++
			}
			flowinfo("main", finished, "of", len(collection_inputs)+len(rule_inputs))
		}

		info("main", "step took", time.Since(startTime))
	}
}

func collectionHandler(name string, input <-chan message, outputs []chan<- message, collection *service.Collection, main chan<- message) {
	c := newCollection(name, collection)

	// timestep loop
	// sends to outputs when "step" is received
	// otherwise, handles operation on the collection
	for {
		flowinfo(name, "ready")
		incoming_message := <-input
		switch incoming_message.operation {
		case "step":
			flowinfo(name, "step")
			outgoing_message := message{
				operation:  "",
				collection: *c,
			}
			for _, output := range outputs {
				output <- outgoing_message
			}
			main <- message{operation: "done"}
			flowinfo(name, "done")
		case "<+", "<~":
			operationinfo(name, "merge", incoming_message.collection.rows)
			c.merge(&incoming_message.collection)
		case "<-":
			operationinfo(name, "delete", incoming_message.collection.rows)
			c.delete(&incoming_message.collection)
		case "<+-":
			operationinfo(name, "delete/merge", incoming_message.collection.rows)
			c.delete(&incoming_message.collection)
			c.merge(&incoming_message.collection)
		default:
			info(name, "unhandled operation:", incoming_message.operation)
		}
		monitorinfo(name, "has", c.rows)
	}
}

func ruleHandler(rule_num int, input <-chan message, output chan<- message, service *service.Service, rule *service.Rule, main chan<- message) {
	// timestep loop
	// 1. receive required collections
	// 2. run rule
	// 3. send resulting collection
	// 4. send finished timestep to main
	for {
		flowinfo(rule_num, "ready")
		// create collections
		collections := make(map[string]*collection)

		// wait for all the required collections to arrive
		for len(collections) != len(rule.Requires()) {
			incoming_message := <-input
			collections[incoming_message.collection.name] = &incoming_message.collection
			flowinfo(rule_num, len(collections), "of", len(rule.Requires()))
		}

		// run rule
		result := runRule(collections, service, rule)

		// spend result
		output <- message{
			operation:  rule.Operation,
			collection: *result,
		}
		flowinfo(rule_num, "sent")

		// tell main step is finished
		main <- message{
			operation: "done",
		}
		flowinfo(rule_num, "done")
	}
}
