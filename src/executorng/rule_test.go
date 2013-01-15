package executorng

/*
TO TEST:
	- tuplesFor
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
		ruleHandler{
			number: 0,
			s: service.Parse("two collections",
				"input in [unimportant, key] => [value]"+
					"table keep [key] => [value]"+
					"keep <+ [in.key, in.value]"),
		},
		map[string]map[string]int{
			"in": map[string]int{
				"unimportant": 0,
				"key":         1,
				"value":       2,
			},
			"keep": map[string]int{
				"key":   0,
				"value": 1,
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
			"keep": map[string]int{
				"key":   0,
				"value": 1,
			},
		},
	})

	for _, test := range tests {
		handler := test[0].(ruleHandler)
		expected := test[1].(map[string]map[string]int)

		result := handler.indexes()

		for collectionName, columns := range result {
			for columnName, resultIndex := range columns {
				expectedIndex := expected[collectionName][columnName]
				if resultIndex != expectedIndex {
					t.Errorf("%v:\n-----\nexpected[%v][%v] = %v; got %v",
						handler.s, collectionName, columnName, expectedIndex, resultIndex)
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
