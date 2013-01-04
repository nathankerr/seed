package main

import (
	"service"
	"testing"
	"testing/quick"
)

func TestNumberOfProducts(t *testing.T) {
	lengths := []int{2}
	expected := 2
	products, _ := numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}

	lengths = []int{2, 2}
	expected = 4
	products, _ = numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}

	lengths = []int{1, 3, 5}
	expected = 15
	products, _ = numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}
}

func TestNumberOfProductsQuick(t *testing.T) {
	f := func(lengths []int) bool {
		containsNegative := false
		for _, length := range lengths {
			if length <= 0 {
				containsNegative = true
				break
			}
		}

		products, err := numberOfProducts(lengths)

		if containsNegative {
			return err != nil && err.Error() == "lengths must be non-negative"
		}

		for _, length := range lengths {
			products /= length
		}
		return products == 1
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestIndexesFor(t *testing.T) {
	checkExpectedProductNumberIndexes(
		map[int][]int{
			0: []int{0},
			1: []int{1},
			2: []int{2},
		},
		[]int{3},
		t,
	)

	checkExpectedProductNumberIndexes(
		map[int][]int{
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
		[]int{3, 3},
		t,
	)
}

func checkExpectedProductNumberIndexes(expected map[int][]int, lengths []int, t *testing.T) {
	numberOfProducts, err := numberOfProducts(lengths)
	if err != nil {
		t.Error(err)
	}
	for productNumber := 0; productNumber < numberOfProducts; productNumber++ {
		indexes, err := indexesFor(productNumber, lengths)
		if err != nil {
			t.Error(err)
		}
		if len(indexes) != len(lengths) {
			t.Errorf("%v != %v", len(indexes), len(lengths))
		}
		for i, index := range indexes {
			if index != expected[productNumber][i] {
				t.Errorf("%v != %v for productNumber %v", indexes, expected[productNumber], productNumber)
				break
			}
		}
	}
}

func TestIndexesForQuick(t *testing.T) {
	f := func(productNumber int, lengths []int) bool {
		indexes, err := indexesFor(productNumber, lengths)

		if err != nil {
			switch err.Error() {
			case "productNumber must be non-negative":
				if productNumber < 0 {
					return true
				}
				return false
			case "lengths must be non-negative":
				containsNonPositive := false
				for _, length := range lengths {
					if length < 0 {
						containsNonPositive = true
						break
					}
				}

				if containsNonPositive {
					return true
				}
				return false
			case "productNumber is greater than the number of possible products":
				return true
			default:
				return false
			}
			return false
		}

		// there should be the same number of indexes as lengths
		if len(indexes) != len(lengths) {
			return false
		}

		// reconstruct the product number from the indexes
		expectedProductNumber := 0
		for i, index := range indexes {
			expectedProductNumber += lengths[i] * index
		}

		return expectedProductNumber == productNumber
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestProductNumberFor(t *testing.T) {
	lengths := []int{3, 3}
	expected := map[int][]int{
		0: []int{0, 0},
		1: []int{1, 0},
		2: []int{2, 0},
		3: []int{0, 1},
		4: []int{1, 1},
		5: []int{2, 1},
		6: []int{0, 2},
		7: []int{1, 2},
		8: []int{2, 2},
	}

	for expectedProductNumber, indexes := range expected {
		productNumber, err := productNumberFor(indexes, lengths)
		if err != nil {
			t.Error(err)
		}

		if productNumber != expectedProductNumber {
			t.Errorf("%v != %v for %v", productNumber, expectedProductNumber, indexes)
		}
	}
}

func TestProjection(t *testing.T) {
	service := service.Parse("projection test",
		"input in [unimportant, key] => [value]"+
			"table keep [key] => [value]"+
			"keep <+ [in.key, in.value]")

	collections := make(map[string]*collection)

	collections["in"] = newCollection("in", service.Collections["in"])
	collections["in"].addRows([][]interface{}{
		[]interface{}{1, 2, 3},
		[]interface{}{4, 5, 6},
	},
	)

	result := runRule(collections, service, service.Rules[0])
	// t.Errorf("%#v", result)

	expected := &collection{
		name:    "keep",
		key:     []string{"key"},
		columns: map[string]int{"key": 0, "value": 1},
		rows: map[string][]interface{}{
			"[2]": []interface{}{2, 3},
			"[5]": []interface{}{5, 6},
		},
	}

	if !collectionsAreEquivalent(expected, result) {
		t.Errorf("\nexpected:\t%v\ngot:\t\t%v", expected, result)
	}
}

func TestProduct(t *testing.T) {
	service := service.Parse("projection test",
		"input a [key]"+
			"input b [key]"+
			"table product [a, b]"+
			"product <+ [a.key, b.key]")

	collections := make(map[string]*collection)

	collections["a"] = newCollection("a", service.Collections["a"])
	collections["a"].addRows([][]interface{}{
		[]interface{}{1},
		[]interface{}{2},
		[]interface{}{3},
	},
	)

	collections["b"] = newCollection("b", service.Collections["b"])
	collections["b"].addRows([][]interface{}{
		[]interface{}{4},
		[]interface{}{5},
		[]interface{}{6},
	},
	)

	expected := &collection{
		name:    "product",
		key:     []string{"a", "b"},
		columns: map[string]int{"a": 0, "b": 1},
		rows: map[string][]interface{}{
			"[1,4]": []interface{}{1, 4},
			"[1,5]": []interface{}{1, 5},
			"[1,6]": []interface{}{1, 6},
			"[2,4]": []interface{}{2, 4},
			"[2,5]": []interface{}{2, 5},
			"[2,6]": []interface{}{2, 6},
			"[3,4]": []interface{}{3, 4},
			"[3,5]": []interface{}{3, 5},
			"[3,6]": []interface{}{3, 6},
		},
	}

	result := runRule(collections, service, service.Rules[0])

	if !collectionsAreEquivalent(expected, result) {
		t.Errorf("\nexpected:\t%v\ngot:\t\t%v", expected, result)
	}
}

func TestFilteredProduct(t *testing.T) {
	service := service.Parse("projection test",
		"input a [key]"+
			"input b [key]"+
			"table intersection [both]"+
			"intersection <+ [a.key]: a.key => b.key")

	collections := make(map[string]*collection)

	collections["a"] = newCollection("a", service.Collections["a"])
	collections["a"].addRows([][]interface{}{
		[]interface{}{1},
		[]interface{}{2},
		[]interface{}{3},
		[]interface{}{4},
		[]interface{}{5},
		[]interface{}{6},
	},
	)

	collections["b"] = newCollection("b", service.Collections["b"])
	collections["b"].addRows([][]interface{}{
		[]interface{}{4},
		[]interface{}{5},
		[]interface{}{6},
	},
	)

	expected := &collection{
		name:    "intersection",
		key:     []string{"both"},
		columns: map[string]int{"both": 0},
		rows: map[string][]interface{}{
			"[4]": []interface{}{4},
			"[5]": []interface{}{5},
			"[6]": []interface{}{6},
		},
	}

	result := runRule(collections, service, service.Rules[0])

	if !collectionsAreEquivalent(expected, result) {
		t.Errorf("\nexpected:\t%v\ngot:\t\t%v", expected, result)
	}
}

// determines if two collections are equivalent
func collectionsAreEquivalent(a, b *collection) bool {
	// check the names
	// string
	if a.name != b.name {
		return false
	}

	// check the keys
	// []string
	if len(a.key) != len(b.key) {
		return false
	}
	for i, value := range a.key {
		if b.key[i] != value {
			return false
		}
	}

	// check the columns
	// map[string]int
	if len(a.columns) != len(b.columns) {
		return false
	}
	for columnName, index := range a.columns {
		bindex, ok := b.columns[columnName]
		if !ok {
			return false
		}

		if index != bindex {
			return false
		}
	}

	// check the rows
	// map[string][]interface{}
	if len(b.rows) != len(a.rows) {
		return false
	}
	for key, row := range a.rows {
		bRow, ok := b.rows[key]
		if !ok {
			return false
		}

		// []interface{}
		if len(bRow) != len(row) {
			return false
		}
		for i, column := range row {
			if bRow[i] != column {
				return false
			}
		}
	}

	return true
}
