package executor

/*
TO TEST:
	- validateData
	- getRequiredData
	- run
	- handleRule
*/

import (
	"encoding/json"
	service "github.com/nathankerr/seed"
	"testing"
)

func TestIndexes(t *testing.T) {
	tests := [][]interface{}{}

	// two collections, both included
	tests = append(tests, []interface{}{
		ruleHandler{ // hander
			number: 0,
			s: service.Parse("two collections",
				"input in [unimportant, key] => [value]"+
					"table keep [key] => [value]"+
					"keep <+ [in.key, in.value]"),
		},
		map[string]map[string]int{ // expected
			"in": map[string]int{
				"unimportant": 0,
				"key":         1,
				"value":       2,
			},
		},
	})

	// one collection, included
	tests = append(tests, []interface{}{
		ruleHandler{
			number: 0,
			s: service.Parse("one collection",
				"table keep [key] => [value]"+
					"keep <+ [keep.key, keep.value]"),
		},
		map[string]map[string]int{
			"keep": map[string]int{
				"key":   0,
				"value": 1,
			},
		},
	})

	// three collections, only two included
	tests = append(tests, []interface{}{
		ruleHandler{
			number: 0,
			s: service.Parse("three collections",
				"input in [unimportant, key] => [value]"+
					"table keep [key] => [value]"+
					"table other [key] => [value]"+
					"keep <+ [in.key, in.value]"),
		},
		map[string]map[string]int{
			"in": map[string]int{
				"unimportant": 0,
				"key":         1,
				"value":       2,
			},
		},
	})

	for _, test := range tests {
		handler := test[0].(ruleHandler)
		expected := test[1].(map[string]map[string]int)

		result := handler.indexes()

		// compare lengths
		expectedLength := len(expected)
		resultLength := len(result)
		if resultLength != expectedLength {
			t.Errorf("expected length of %v, got %v for %v and %v",
				expectedLength, resultLength, expected, result)
		}

		// compare contents
		for expectedCollectionName, expectedColumns := range expected {
			resultColumns, ok := result[expectedCollectionName]
			if !ok {
				t.Errorf("expected collection name, %v, does not exist in result %v",
					expectedCollectionName, result)
			}

			// compare column contents
			for expectedColumnName, expectedColumnIndex := range expectedColumns {
				resultColumnIndex, ok := resultColumns[expectedColumnName]
				if !ok {
					t.Errorf("expected column name, %v, does not exist in result columns %v",
						expectedColumnName, resultColumns)
				}

				if resultColumnIndex != expectedColumnIndex {
					t.Errorf("expected %v, got %v for %v in %v and %v",
						expectedColumnIndex, resultColumnIndex, expectedColumnName, expectedColumns, resultColumns)
				}
			}
		}
	}
}

func TestNumberOfProducts(t *testing.T) {
	tests := [][]interface{}{}

	// one
	tests = append(tests, []interface{}{
		[]int{2}, // lengths
		2,        // expected	
	})

	// two
	tests = append(tests, []interface{}{
		[]int{2, 2}, // lengths
		4,           // expected	
	})

	// three
	tests = append(tests, []interface{}{
		[]int{1, 3, 5}, // lengths
		15,             // expected	
	})

	for _, test := range tests {
		lengths := test[0].([]int)
		expected := test[1].(int)

		result := numberOfProducts(lengths)

		if result != expected {
			t.Errorf("expected numberOfProducts(%v)=%v; got %v",
				lengths, expected, result)
		}
	}
}

func TestProductNumberFor(t *testing.T) {
	tests := [][]interface{}{}

	// 2x2
	tests = append(tests, []interface{}{
		[]int{2, 2}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{1, 0},
			2: []int{0, 1},
			3: []int{1, 1},
		},
	})

	// 3x3
	tests = append(tests, []interface{}{
		[]int{3, 3}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{1, 0},
			2: []int{2, 0},
			3: []int{0, 1},
			4: []int{1, 1},
			5: []int{2, 1},
			6: []int{0, 2},
			7: []int{1, 2},
			8: []int{2, 2},
		},
	})

	for _, test := range tests {
		lengths := test[0].([]int)
		expected := test[1].(map[int][]int)

		for expectedProductNumber, indexes := range expected {
			resultProductNumber := productNumberFor(indexes, lengths)

			if resultProductNumber != expectedProductNumber {
				t.Errorf("productNumberFor(%v, %v)=%v; expected %v",
					indexes, lengths, expectedProductNumber, resultProductNumber)
			}
		}
	}
}

func TestIndexesFor(t *testing.T) {
	tests := [][]interface{}{}

	// 1x1
	tests = append(tests, []interface{}{
		[]int{1, 1}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
		},
	})

	// 1x2
	tests = append(tests, []interface{}{
		[]int{1, 2}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{0, 1},
		},
	})

	// 2x2
	tests = append(tests, []interface{}{
		[]int{2, 2}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{1, 0},
			2: []int{0, 1},
			3: []int{1, 1},
		},
	})

	// 2x3
	tests = append(tests, []interface{}{
		[]int{2, 3}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{1, 0},
			2: []int{0, 1},
			3: []int{1, 1},
			4: []int{0, 2},
			5: []int{1, 2},
		},
	})

	// 3x2
	tests = append(tests, []interface{}{
		[]int{3, 2}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{1, 0},
			2: []int{2, 0},
			3: []int{0, 1},
			4: []int{1, 1},
			5: []int{2, 1},
		},
	})

	// 3x3
	tests = append(tests, []interface{}{
		[]int{3, 3}, // lengths
		map[int][]int{ // expected
			0: []int{0, 0},
			1: []int{1, 0},
			2: []int{2, 0},
			3: []int{0, 1},
			4: []int{1, 1},
			5: []int{2, 1},
			6: []int{0, 2},
			7: []int{1, 2},
			8: []int{2, 2},
		},
	})

	for _, test := range tests {
		lengths := test[0].([]int)
		expected := test[1].(map[int][]int)

		for productNumber, expectedIndexes := range expected {
			result := indexesFor(productNumber, lengths)

			for i, expectedIndex := range expectedIndexes {
				resultIndex := result[i]

				if resultIndex != expectedIndex {
					t.Errorf("expected %v, got %v for %v and %v",
						expectedIndex, resultIndex, expectedIndexes, result)
				}
			}
		}
	}
}

func TestTuplesFor(t *testing.T) {
	tests := [][]interface{}{}

	tests = append(tests, []interface{}{
		map[string][]tuple{ // data
			"a": []tuple{
				tuple{1, 2},
				tuple{3, 4},
			},
			"b": []tuple{
				tuple{5, 6},
				tuple{7, 8},
			},
		},
		[]string{ // collections
			"a",
			"b",
		},
		map[int]map[string]tuple{ // expected[productNumber][collection]
			0: map[string]tuple{
				"a": tuple{1, 2},
				"b": tuple{5, 6},
			},
			1: map[string]tuple{
				"a": tuple{3, 4},
				"b": tuple{5, 6},
			},
		},
	})

	// same as for product test in TestCalculateResults
	tests = append(tests, []interface{}{
		map[string][]tuple{ // data
			"a": []tuple{
				tuple{1},
				tuple{2},
				tuple{3},
			},
			"b": []tuple{
				tuple{4},
				tuple{5},
				tuple{6},
			},
		},
		[]string{ //collections
			"a",
			"b",
		},
		map[int]map[string]tuple{ // expected[productNumber][collection]
			0: map[string]tuple{
				"a": tuple{1},
				"b": tuple{4},
			},
			1: map[string]tuple{
				"a": tuple{2},
				"b": tuple{4},
			},
			2: map[string]tuple{
				"a": tuple{3},
				"b": tuple{4},
			},
			3: map[string]tuple{
				"a": tuple{1},
				"b": tuple{5},
			},
			4: map[string]tuple{
				"a": tuple{2},
				"b": tuple{5},
			},
			5: map[string]tuple{
				"a": tuple{3},
				"b": tuple{5},
			},
			6: map[string]tuple{
				"a": tuple{1},
				"b": tuple{6},
			},
			7: map[string]tuple{
				"a": tuple{2},
				"b": tuple{6},
			},
			8: map[string]tuple{
				"a": tuple{3},
				"b": tuple{6},
			},
		},
	})

	for _, test := range tests {
		data := test[0].(map[string][]tuple)
		collections := test[1].([]string)
		expectedProducts := test[2].(map[int]map[string]tuple)

		lengths := []int{}
		for _, collection := range collections {
			tuples := data[collection]
			lengths = append(lengths, len(tuples))
		}

		for productNumber, expected := range expectedProducts {
			result := tuplesFor(productNumber, collections, lengths, data)

			// compare lengths
			resultLength := len(result)
			expectedLength := len(expected)
			if resultLength != expectedLength {
				t.Errorf("expected length of %v, got %v for %v and %v",
					expectedLength, resultLength, expected, result)
			}

			// compare tuples
			for expectedCollectionName, expectedTuple := range expected {
				resultTuple, ok := result[expectedCollectionName]
				if !ok {
					t.Errorf("expected a collection named %v, result does not have one",
						expectedCollectionName)
				}

				// compare tuple lengths
				expectedTupleLength := len(expectedTuple)
				resultTupleLength := len(resultTuple)
				if resultTupleLength != expectedTupleLength {
					t.Errorf("expected tuple length of %v, got %v",
						expectedTupleLength, resultTupleLength)
				}

				// compare tuple contents
				equal := true
				for columnIndex, expectedColumn := range expectedTuple {
					resultColumn := resultTuple[columnIndex]
					if resultColumn != expectedColumn {
						equal = false
					}
				}
				if !equal {
					t.Errorf("expected %v, got %v from %v and %v for productNumber %v",
						expectedTuple, resultTuple, expected, result, productNumber)
				}
			}
		}
	}
}

func TestCalculateResults(t *testing.T) {
	tests := [][]interface{}{}

	// projection
	tests = append(tests, []interface{}{
		ruleHandler{ // handler
			number: 0,
			s: service.Parse("projection test",
				"input in [unimportant, key] => [value]"+
					"table keep [key] => [value]"+
					"keep <+ [in.key, in.value]"),
			//channels: ,
		},
		map[string][]tuple{ // data
			"in": []tuple{
				tuple{1, 2, 3},
				tuple{4, 5, 6},
			},
		},
		[]tuple{ // expected
			tuple{2, 3},
			tuple{5, 6},
		},
	})

	// product
	tests = append(tests, []interface{}{
		ruleHandler{ // handler
			number: 0,
			s: service.Parse("product test",
				"input a [key]"+
					"input b [key]"+
					"table keep [key]"+
					"keep <+ [a.key, b.key]"),
			//channels: ,
		},
		map[string][]tuple{ // data
			"a": []tuple{
				tuple{1},
				tuple{2},
				tuple{3},
			},
			"b": []tuple{
				tuple{4},
				tuple{5},
				tuple{6},
			},
		},
		[]tuple{ // expected
			tuple{1, 4},
			tuple{2, 4},
			tuple{3, 4},
			tuple{1, 5},
			tuple{2, 5},
			tuple{3, 5},
			tuple{1, 6},
			tuple{2, 6},
			tuple{3, 6},
		},
	})

	// filter
	tests = append(tests, []interface{}{
		ruleHandler{ // handler
			number: 0,
			s: service.Parse("filter test",
				"input a [key]"+
					"input b [key]"+
					"table intersection [both]"+
					"intersection <+ [a.key]: a.key => b.key"),
			//channels: ,
		},
		map[string][]tuple{ // data
			"a": []tuple{
				tuple{1},
				tuple{2},
				tuple{3},
				tuple{4},
				tuple{5},
				tuple{6},
			},
			"b": []tuple{
				tuple{4},
				tuple{5},
				tuple{6},
			},
		},
		[]tuple{ // expected
			tuple{4},
			tuple{5},
			tuple{6},
		},
	})

	for _, test := range tests {
		handler := test[0].(ruleHandler)
		data := test[1].(map[string][]tuple)
		expected := test[2].([]tuple)

		result := handler.calculateResults(data)

		// compare lengths
		expectedLength := len(expected)
		resultLength := len(result)
		if resultLength != expectedLength {
			t.Errorf("expected length of %v, got %v for %v and %v",
				expectedLength, resultLength, expected, result)
		}

		// turn the list of result tuples into a map
		resultTuples := map[string]tuple{}
		for _, tuple := range expected {
			jsonified, err := json.Marshal(tuple)
			if err != nil {
				panic(err)
			}

			resultTuples[string(jsonified)] = tuple
		}

		// check to see if the expected tuples are in the result
		for _, expectedTuple := range expected {
			jsonifiedExpectedTuple, err := json.Marshal(expectedTuple)
			if err != nil {
				panic(err)
			}

			_, ok := resultTuples[string(jsonifiedExpectedTuple)]
			if !ok {
				t.Errorf("%v not found in result %v",
					expectedTuple, result)
			}
		}
	}
}
