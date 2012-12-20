package main

import(
	"fmt"
	"io/ioutil"
	"service"
	"time"
)

type input struct {
	channel string
	data map[string]string
}

type message struct {
	collection string
	data map[string]string
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

	input_message := input{channel: "kvget", data: map[string]string{"key": "123"}}

	collection_inputs["kvget"] <- message{data: input_message.data}
	collection_inputs["kvget"] <- message{data: map[string]string{"key": "456"}}
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