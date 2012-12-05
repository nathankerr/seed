package main

import (
	"encoding/json"
)

func (s *service) toJson(name string) string {
	info()
	marshaled, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(marshaled)
}

func (ct collectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}
