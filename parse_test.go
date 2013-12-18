package seed

import (
	"testing"
)

func TestParser(t *testing.T) {
	input := `input request [timezone]
output response [timezone, current_time]

response <+ [request.timezone, (current_time_in request.timezone)]`

	expected := &Seed{
		Name: "time",
		Source: Source{
			Name:   "time",
			Line:   1,
			Column: 1,
		},
		Collections: map[string]*Collection{
			"response": &Collection{
				Type: CollectionOutput,
				Key:  []string{"timezone", "current_time"},
				Data: []string(nil),
				Source: Source{
					Name:   "time",
					Line:   2,
					Column: 8,
				},
			},
			"request": &Collection{
				Type: CollectionInput,
				Key:  []string{"timezone"},
				Data: []string(nil),
				Source: Source{
					Name:   "time",
					Line:   1,
					Column: 6,
				},
			},
		},
		Rules: []*Rule{
			&Rule{
				Supplies:  "response",
				Operation: "<+",
				Projection: []Expression{
					Expression{Value: QualifiedColumn{
						Collection: "request",
						Column:     "timezone",
					}},
					Expression{Value: MapFunction{
						Name: "current_time_in",
						// Function: current_time_in,
						Arguments: []QualifiedColumn{
							QualifiedColumn{
								Collection: "request",
								Column:     "timezone",
							}},
					}},
				},
				// Predicate: []Constraint{},
				Source: Source{
					Name:   "time",
					Line:   4,
					Column: 1,
				},
			},
		},
	}

	output := Parse("time", input)

	err := output.EquivalentTo(expected)
	if err != nil {
		t.Error(err)
	}
}
