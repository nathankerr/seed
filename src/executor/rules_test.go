package main

import(
	"testing"
	"testing/quick"
)

func TestNumberOfProducts (t *testing.T) {
	lengths := []int{2}
	expected := 2
	products, _ := numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}

	lengths = []int{2,2}
	expected = 4
	products, _ = numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}

	lengths = []int{1,3,5}
	expected = 15
	products, _ = numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}
}

func TestNumberOfProductsQuick (t *testing.T) {
	f := func(lengths []int) bool {
		containsNonPositive := false
		for _, length := range lengths {
			if length <= 0 {
				containsNonPositive = true
				break
			}
		}

		products, err := numberOfProducts(lengths)

		if containsNonPositive {
			return err != nil && err.Error() == "lengths must be positive"
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

func TestIndexesFor (t *testing.T) {
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

func TestIndexesForQuick (t *testing.T) {
	f := func(productNumber int, lengths []int) bool {
		indexes, err := indexesFor(productNumber, lengths)

		if err != nil {
			switch err.Error() {
			case "productNumber must be non-negative":
				if productNumber < 0 {
					return true
				}
				return false
			case "lengths must be positive":
				containsNonPositive := false
				for _, length := range lengths {
					if length <= 0 {
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

func TestProductNumberFor (t *testing.T) {
	lengths := []int{3,3}
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