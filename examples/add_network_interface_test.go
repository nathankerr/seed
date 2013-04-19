package examples

import (
	"github.com/nathankerr/seed"
	"testing"
	// "reflect"
)

func TestGroupTypes(t *testing.T) {
	type test struct {
		name         string
		input        string
		expectedName string
		expected     *seed.Seed
	}

	tests := []test{
		test{
			name: "time",
			input: `input request [timezone]
output response [timezone, current_time]

response <+ [request.timezone, (current_time_in request.timezone)]`,
			expected: &seed.Seed{
				Name: "TimeServer",
				Source: seed.Source{
					Name:   "time",
					Line:   1,
					Column: 1,
				},
				Collections: map[string]*seed.Collection{
					"response": &seed.Collection{
						Type: seed.CollectionChannel,
						Key:  []string{"@response_addr", "timezone", "current_time"},
						Data: []string(nil),
						Source: seed.Source{
							Name:   "time",
							Line:   2,
							Column: 8,
						},
					},
					"request": &seed.Collection{
						Type: seed.CollectionChannel,
						Key:  []string{"@address", "response_addr", "timezone"},
						Data: []string(nil),
						Source: seed.Source{
							Name:   "time",
							Line:   1,
							Column: 6,
						},
					},
				},
				Rules: []*seed.Rule{
					&seed.Rule{
						Supplies:  "response",
						Operation: "<~",
						Projection: []seed.Expression{
							seed.Expression{Value: seed.QualifiedColumn{
								Collection: "request",
								Column:     "response_addr",
							}},
							seed.Expression{Value: seed.QualifiedColumn{
								Collection: "request",
								Column:     "timezone",
							}},
							seed.Expression{Value: seed.MapFunction{
								Name: "current_time_in",
								// Function: current_time_in,
								Arguments: []seed.QualifiedColumn{
									seed.QualifiedColumn{
										Collection: "request",
										Column:     "timezone",
									}},
							}},
						},
						// Predicate: []seed.Constraint{},
						Source: seed.Source{
							Name:   "time",
							Line:   4,
							Column: 1,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		input := seed.Parse(test.name, test.input)
		// expected := seed.Parse(test.name, test.expected, false)
		expected := test.expected

		output, err := Add_network_interface(input)
		if err != nil {
			t.Errorf("%v: %v", test.name, err)
		}

		err = expected.EquivalentTo(output)
		if err != nil {
			t.Errorf("[%v] %v", test.name, err)
		}

		// if !reflect.DeepEqual(expected, output) {
		// 	t.Errorf("%v: expected %#v, got %#v", test.name, expected, output)
		// }

	}
}
