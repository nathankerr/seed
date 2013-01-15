package executorng

/*
TO TEST:
	- validateData
	- calculateResults
	- getRequiredData
	- run
	- handleRule
*/

import (
	"service"
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
			t.Errorf("expected numberOfProducts(%v)=%v; got",
				lengths, expected, result)
		}
	}
}

func TestProductNumberFor(t *testing.T) {
	tests := [][]interface{}{}

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
					t.Errorf("expected %v, got %v",
						expectedIndex, resultIndex)
				}
			}
		}
	}
}

func TestTuplesFor(t *testing.T) {
	tests := [][]interface{}{}

	tests = append(tests, []interface{}{
		0, //productNumber
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
		map[string]tuple{ // expected
			"a": tuple{1, 2},
			"b": tuple{5, 6},
		},
	})

	for _, test := range tests {
		productNumber := test[0].(int)
		data := test[1].(map[string][]tuple)
		expected := test[2].(map[string]tuple)

		result := tuplesFor(productNumber, data)

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
			for columnIndex, expectedColumn := range expectedTuple {
				resultColumn := resultTuple[columnIndex]
				if resultColumn != expectedColumn {
					t.Errorf("expected %v, got %v",
						expectedColumn, resultColumn)
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

		// compare tuples
		for tupleIndex, expectedTuple := range expected {
			resultTuple := result[tupleIndex]

			// compare tuple lengths
			expectedTupleLength := len(expectedTuple)
			resultTupleLength := len(resultTuple)
			if resultTupleLength != expectedTupleLength {
				t.Errorf("expected length of %v, got %v for %v and %v",
					expectedTupleLength, resultTupleLength, expectedTuple, resultTuple)
			}

			// compare tuple contents
			for columnIndex, expectedColumn := range expectedTuple {
				resultColumn := resultTuple[columnIndex]

				if resultColumn != expectedColumn {
					t.Errorf("expected %v, got %v in comparing %v and %v",
						expectedColumn, resultColumn, expectedTuple, resultTuple)
				}
			}
		}
	}
}
