package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nathankerr/seed"
	"reflect"
)

type ruleHandler struct {
	number   int
	s        *seed.Seed
	channels Channels
}

func handleRule(ruleNumber int, s *seed.Seed, channels Channels) {
	controlinfo(ruleNumber, "started")
	handler := ruleHandler{
		number:   ruleNumber,
		s:        s,
		channels: channels,
	}

	input := channels.Rules[ruleNumber]
	rule := s.Rules[ruleNumber]
	dataMessages := []MessageContainer{}

	for {
		message := <-input
		controlinfo(ruleNumber, "received", message)

		switch message.Operation {
		case "immediate":
			var results MessageContainer
			if rule.Operation == "<=" {
				results = handler.run(dataMessages)
			}
			dataMessages = []MessageContainer{}
			results.Operation = "done"
			results.Collection = fmt.Sprint(handler.number)
			channels.Control <- results
			controlinfo(ruleNumber, "finished with", message)
		case "deferred":
			var results MessageContainer
			if rule.Operation != "<=" {
				results = handler.run(dataMessages)
			}
			dataMessages = []MessageContainer{}
			results.Operation = "done"
			results.Collection = fmt.Sprint(handler.number)
			channels.Control <- results
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

func (handler *ruleHandler) run(dataMessages []MessageContainer) MessageContainer {
	// get the data needed to calculate the results
	data := handler.getRequiredData(dataMessages)

	// calculate results
	results := handler.calculateResults(data)

	// send results
	outputName := handler.s.Rules[handler.number].Supplies
	outputMessage := MessageContainer{
		Operation:  "data",
		Collection: outputName,
		Data:       results,
	}
	handler.channels.Collections[outputName] <- outputMessage
	flowinfo(handler.number, "sent", outputMessage.String(), "to", outputName)

	return outputMessage
}

func (handler *ruleHandler) getRequiredData(dataMessages []MessageContainer) map[string][]seed.Tuple {
	data := map[string][]seed.Tuple{}

	// process cached data
	for _, message := range dataMessages {
		data[message.Collection] = message.Data
	}

	// receive other needed data
	required := len(handler.s.Rules[handler.number].Requires())
	input := handler.channels.Rules[handler.number]
	for stillNeeded := required - len(dataMessages); stillNeeded > 0; stillNeeded-- {
		message := <-input
		controlinfo(handler.number, "received", message)

		switch message.Operation {
		case "data":
			flowinfo(handler.number, "received", message.String())
			data[message.Collection] = message.Data
		default:
			fatal(handler.number, "unhandled message", message)
		}
	}

	if _, ok := data[""]; ok {
		fatal(handler.number, "received data without a collection name")
	}

	return data
}

func (handler *ruleHandler) calculateResults(data map[string][]seed.Tuple) []seed.Tuple {
	// validate data and handle errors
	err := handler.validateData(data)
	if err != nil {
		switch err.Error() {
		case "empty tuple set":
			return []seed.Tuple{}
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
	results := map[string]seed.Tuple{}
	reductions := map[string]map[int][]seed.Tuple{} // json-ified version of row to which the reduction will be used for
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
		result := seed.Tuple{}
		localReductions := map[int]seed.Tuple{} // column number: arguments
		for columnNumber, expression := range rule.Projection {
			var element interface{}

			// determine the element to add to the result tuple
			switch value := expression.(type) {
			case seed.QualifiedColumn:
				columnIndex := indexes[value.Collection][value.Column]
				element = tuples[value.Collection][columnIndex]
			case seed.MapFunction:
				// gather arguments
				arguments := seed.Tuple{}
				for _, qc := range value.Arguments {
					columnIndex := indexes[qc.Collection][qc.Column]
					arguments = append(arguments, tuples[qc.Collection][columnIndex])
				}

				// run function to get result
				element = value.Function(arguments)
			case seed.ReduceFunction:
				// add a place holder to the result tuple
				element = nil

				// gather the arguments for this reduction
				arguments := seed.Tuple{}
				for _, qc := range value.Arguments {
					columnIndex := indexes[qc.Collection][qc.Column]
					arguments = append(arguments, tuples[qc.Collection][columnIndex])
				}
				localReductions[columnNumber] = arguments
			default:
				panic(fmt.Sprintf("unhandled type: %v", reflect.TypeOf(expression).String()))
			}

			result = append(result, element)
		}

		setidBytes, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		setid := string(setidBytes)

		// add the result to the set of results
		results[setid] = result

		// add local reductions to list of all reductions (by group)
		for columnNumber, reductionTuple := range localReductions {
			if _, ok := reductions[setid]; !ok {
				reductions[setid] = map[int][]seed.Tuple{}
			}
			reductions[setid][columnNumber] = append(reductions[setid][columnNumber], reductionTuple)
		}
	}

	// run reductions and add results to the appropriate places before returning the results
	resultsSlice := make([]seed.Tuple, 0, len(results))
	for setid, result := range results {
		for columnNumber, reductionTuples := range reductions[setid] {
			result[columnNumber] = rule.Projection[columnNumber].(seed.ReduceFunction).Function(reductionTuples)
		}
		resultsSlice = append(resultsSlice, result)
	}
	return resultsSlice
}

func (handler *ruleHandler) validateData(data map[string][]seed.Tuple) error {
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

func tuplesFor(productNumber int, collections []string, lengths []int, data map[string][]seed.Tuple) map[string]seed.Tuple {
	tupleIndexes := indexesFor(productNumber, lengths)
	tuples := map[string]seed.Tuple{}
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
