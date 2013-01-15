package executorng

import(
	"testing"
	"service"
)

func TestNumberOfProducts(t *testing.T) {
	lengths := []int{2}
	expected := 2
	products := numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}

	lengths = []int{2, 2}
	expected = 4
	products = numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}

	lengths = []int{1, 3, 5}
	expected = 15
	products = numberOfProducts(lengths)
	if products != expected {
		t.Errorf("%v != %v", expected, products)
	}
}

func TestProjection(t *testing.T) {
	service := service.Parse("projection test",
		"input in [unimportant, key] => [value]"+
			"table keep [key] => [value]"+
			"keep <+ [in.key, in.value]")

	handler := ruleHandler{
		number: 0,
		s: service,
	}

	data := map[string][]tuple{
		"in": []tuple{
			tuple{1, 2, 3},
			tuple{4, 5, 6},
		},
	}

	result := handler.calculateResults(data)

	expected := []tuple{
		tuple{2, 3},
		tuple{5, 6},
	}
	
	if !tuplesAreEquivalent(expected, result) {
		t.Errorf("\nexpected:\t%v\ngot:\t\t%v", expected, result)
	}
}

func tuplesAreEquivalent(results []tuple, expected []tuple) bool {
	if len(results) != len(expected) {
		return false
	}

	for tupleNum := 0; tupleNum < len(results); tupleNum++ {
		resultTuple := results[tupleNum]
		expectedTuple := expected[tupleNum]

		if len(resultTuple) != len(expectedTuple) {
			return false
		}

		for columnNum := 0; columnNum < len(resultTuple); columnNum++ {
			resultColumn := resultTuple[columnNum]
			expectedColumn := expectedTuple[columnNum]
			if resultColumn != expectedColumn {
				return false
			}
		}
	}

	return true
}