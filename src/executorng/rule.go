package executorng

import (
	"errors"
	"service"
)

type ruleHandler struct {
	number   int
	s        *service.Service
	channels channels
}

func handleRule(ruleNumber int, s *service.Service, channels channels) {
	controlinfo(ruleNumber, "started")
	handler := ruleHandler{
		number:   ruleNumber,
		s:        s,
		channels: channels,
	}

	input := channels.rules[ruleNumber]
	rule := s.Rules[ruleNumber]
	dataMessages := []messageContainer{}

	for {
		message := <-input
		controlinfo(ruleNumber, "received", message)

		switch message.operation {
		case "immediate":
			if rule.Operation == "<=" {
				handler.run(dataMessages)
			}
			dataMessages = []messageContainer{}
			channels.finished <- true
			controlinfo(ruleNumber, "finished with", message)
		case "deferred":
			if rule.Operation != "<=" {
				handler.run(dataMessages)
			}
			dataMessages = []messageContainer{}
			channels.finished <- true
			controlinfo(ruleNumber, "finished with", message)
		case "data":
			// cache data received before an immediate or deferred message initiates execution
			flowinfo(handler.number, "received", message.String())
			dataMessages = append(dataMessages, message)
		default:
			fatal(ruleNumber, "unhandled message:", message)
		}
	}
}

func (handler *ruleHandler) run(dataMessages []messageContainer) {
	// get the data needed to calculate the results
	data := handler.getRequiredData(dataMessages)

	// calculate results
	results := handler.calculateResults(data)

	// send results
	outputName := handler.s.Rules[handler.number].Supplies
	outputMessage := messageContainer{
		operation:  "data",
		collection: outputName,
		data:       results,
	}
	handler.channels.collections[outputName] <- outputMessage
	flowinfo(handler.number, "sent", outputMessage.String(), "to", outputName)
}

func (handler *ruleHandler) getRequiredData(dataMessages []messageContainer) map[string][]tuple {
	data := map[string][]tuple{}

	// process cached data
	for _, message := range dataMessages {
		data[message.collection] = message.data
	}

	// receive other needed data
	required := len(handler.s.Rules[handler.number].Requires())
	input := handler.channels.rules[handler.number]
	for stillNeeded := required - len(dataMessages); stillNeeded > 0; stillNeeded-- {
		message := <-input
		controlinfo(handler.number, "received", message)

		switch message.operation {
		case "data":
			flowinfo(handler.number, "received", message.String())
			data[message.collection] = message.data
		default:
			fatal(handler.number, "unhandled message", message)
		}
	}

	if _, ok := data[""]; ok {
		fatal(handler.number, "received data without a collection name")
	}

	return data
}

func (handler *ruleHandler) calculateResults(data map[string][]tuple) []tuple {
	// validate data and handle errors
	err := handler.validateData(data)
	if err != nil {
		switch err.Error() {
		case "empty tuple set":
			return []tuple{}
		default:
			fatal(handler.number, err)
		}
	}

	// get indexes for resolving collection.column references
	indexes := handler.indexes()

	// get the number of products
	collections := []string{}
	lengths := []int{}
	for collection, tuples := range data {
		collections = append(collections, collection)
		lengths = append(lengths, len(tuples))
	}
	numberOfProducts := numberOfProducts(lengths)

	rule := handler.s.Rules[handler.number]
	results := []tuple{}
	for productNumber := 0; productNumber < numberOfProducts; productNumber++ {
		// get the tuples for this product
		tuples := tuplesFor(productNumber, collections, lengths, data)

		// skip this product if the predicate is not fulfilled
		skip := true
		if len(rule.Predicate) == 0 {
			skip = false
		}
		for _, constraint := range rule.Predicate {
			// get the left column
			lqc := constraint.Left
			leftColumnIndex := indexes[lqc.Collection][lqc.Column]
			left := tuples[lqc.Collection][leftColumnIndex]

			// get the right column
			rqc := constraint.Right
			rightColumnIndex := indexes[rqc.Collection][rqc.Column]
			right := tuples[rqc.Collection][rightColumnIndex]

			if left == right {
				skip = false
			}
		}
		if skip {
			continue
		}

		// generate the result row and add to the set of results
		result := tuple{}
		for _, qc := range rule.Projection {
			columnIndex := indexes[qc.Collection][qc.Column]
			result = append(result, tuples[qc.Collection][columnIndex])
		}
		results = append(results, result)
	}

	return results
}

func (handler *ruleHandler) validateData(data map[string][]tuple) error {
	for collectionName, tuples := range data {
		collection := handler.s.Collections[collectionName]

		// each set of tuples should contain tuples
		if len(tuples) < 1 {
			return errors.New("empty tuple set")
		}

		// each tuple should have the correct length
		correctLength := len(collection.Key) + len(collection.Data)
		for _, tuple := range tuples {
			if len(tuple) != correctLength {
				return errors.New("tuple has the wrong length")
			}
		}
	}

	return nil
}

func numberOfProducts(lengths []int) int {
	products := 1
	for _, length := range lengths {
		if length != 0 {
			products *= length
		}
	}
	return products
}

func tuplesFor(productNumber int, collections []string, lengths []int, data map[string][]tuple) map[string]tuple {
	tupleIndexes := indexesFor(productNumber, lengths)
	tuples := map[string]tuple{}
	for nameIndex, tupleIndex := range tupleIndexes {
		collectionName := collections[nameIndex]
		tuples[collectionName] = data[collectionName][tupleIndex]
	}
	return tuples
}

func indexesFor(productNumber int, lengths []int) []int {
	// this is sort of a reverse base conversion which starts with
	// the least significant digit
	// indexes holds the digits, indexes[0] is least significant
	indexes := []int{}
	for i, length := range lengths {
		index := productNumber

		// find out how many products have already been
		// represented and remove them from consideration by
		// this index
		productsBeforeThis := productNumberFor(indexes, lengths[:i])
		index -= productsBeforeThis

		// now the index is just a factor off
		// find the factor and remove it
		factor := numberOfProducts(lengths[:i])
		index /= factor

		// now limit the index to what it can hold (length)
		// any remaining part of productNumber will be
		// handled by the next index
		if index >= length {
			if length == 0 {
				index = 0
			} else {
				index %= length
			}
		}

		indexes = append(indexes, index)
	}

	return indexes
}

// given a set of indexes, return the product number
func productNumberFor(indexes []int, lengths []int) int {
	productNumber := 0
	for i, _ := range indexes {
		factor := numberOfProducts(lengths[:i])
		productNumber += factor * indexes[i]
	}

	return productNumber
}

// returns a map of maps telling what the slice index of a column is
// indexes[collectionName][columnName]
func (handler *ruleHandler) indexes() map[string]map[string]int {
	indexes := map[string]map[string]int{}
	rule := handler.s.Rules[handler.number]

	for _, collectionName := range rule.Requires() {
		collection := handler.s.Collections[collectionName]
		indexes[collectionName] = map[string]int{}

		for index, columnName := range collection.Key {
			indexes[collectionName][columnName] = index
		}

		baseIndex := len(collection.Key)
		for index, columnName := range collection.Data {
			indexes[collectionName][columnName] = baseIndex + index
		}
	}

	return indexes
}
