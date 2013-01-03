package main

import (
	"errors"
)

func productNumberToIndexes(productNumber int, lengths []int) []int {
	return nil
}


func numberOfProducts(lengths []int) (int, error) {
	products := 1
	for _, length := range lengths {
		if length <= 0 {
			return -1, errors.New("lengths must be positive")
		}
		products *= length
	}
	return products, nil
}