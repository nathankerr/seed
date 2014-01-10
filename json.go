package seed

import (
	"encoding/json"
	"errors"
)

func FromJson(name string, input []byte) (*Seed, error) {
	info()

	seed := &Seed{Collections: make(map[string]*Collection)}

	err := json.Unmarshal(input, seed)

	return seed, err
}

func ToJson(seed *Seed, name string) ([]byte, error) {
	info()
	return json.MarshalIndent(seed, "", "\t")
}

func (ct CollectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}

func (ct CollectionType) UnmarshalJSON(input []byte) error {
	// check for "" at beginning and end
	if input[0] != '"' && input[len(input)-1] != '"' {
		panic("not a string")
	}

	switch string(input[1 : len(input)-1]) {
	case "input":
		ct = CollectionInput
	case "output":
		ct = CollectionOutput
	case "table":
		ct = CollectionTable
	case "channel":
		ct = CollectionChannel
	case "scratch":
		ct = CollectionScratch
	default:
		return errors.New("Unknown collection type: " + string(input[1:len(input)-1]))
	}

	return nil
}

// func (t Tuple) UnmarshalJSON(input []byte) error {
// 	info(string(input))
// 	var data []interface{}

// 	err := json.Unmarshal(input, &data)
// 	if err != nil {
// 		return err
// 	}

// 	for _, item := range data {
// 		t = append(t, item)
// 	}

// 	return nil
// }
