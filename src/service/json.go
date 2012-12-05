package service

import (
	"encoding/json"
)

func (s *Service) ToJson(name string) string {
	info()
	marshaled, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(marshaled)
}

func (ct CollectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}
