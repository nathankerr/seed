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
			return err != nil //&& err.Error() == "lengths must be positive"
		} else {
			for _, length := range lengths {
				products /= length
			}
			return products == 1
		}
		return false
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestProductNumberToIndexes (t *testing.T) {
	numberOfProducts := 0
	for productNumber := 0; productNumber < numberOfProducts; productNumber++ {
		// productNumberToIndexes(productNumber int, lengths []int)
	}
}