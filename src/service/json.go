package service

import (
	"encoding/json"
)

func ToJson(seed *Service, name string) ([]byte, error) {
	info()
	return json.MarshalIndent(seed, "", "\t")
}

func (ct CollectionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.String())
}
