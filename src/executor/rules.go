package main

import (
	"errors"
)

func indexesFor(productNumber int, lengths []int) ([]int, error) {
	// productNumber must be non-negative
	if productNumber < 0 {
		return nil, errors.New("productNumber must be non-negative")
	}

	// lengths must be positive
	for _, length := range lengths {
		if length <= 0 {
			return nil, errors.New("lengths must be positive")
		}
	}
	
	// productNumber cannot be greater than the number of possible products
	numberOfProducts1, err := numberOfProducts(lengths)
	if err != nil {
		return nil, err
	}
	if productNumber > numberOfProducts1 {
		return nil, errors.New("productNumber is greater than the number of possible products")
	}

	// this is sort of a reverse base coversion which starts with
	// the least significant digit
	// indexes holds the digits, indexes[0] is least significant
	indexes := make([]int, 0)
	for i, length := range lengths {
		index := productNumber

		// find out how many products have already been
		// represented and remove them from consideration by
		// this index
		productsBeforeThis, err := productNumberFor(indexes, lengths[:i])
		if err != nil {
			panic(err)
		}
		index -= productsBeforeThis

		// now the index is just a factor off
		// find the factor and remove it
		factor, err := numberOfProducts(lengths[:i])
		if err != nil {
			panic(err)
		}
		index /= factor

		// now limit the index to what it can hold (length)
		// any remaining part of productNumber will be
		// handled by the next index
		if index >= length {
			index %= length
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

func productNumberFor(indexes []int, lengths []int) (int, error) {
	// indexes and arrays must be the same length
	if len(indexes) != len(lengths) {
		return -1, errors.New("indexes and lengths must have the same lengths")
	}

	// check ranges for indexes and lengths
	for i, _ := range indexes {
		if indexes[i] < 0 {
			return -1, errors.New("indexes must be non-negative")
		}
		if lengths[i] <= 0 {
			return -1, errors.New("lengths must be positive")
		}
	}

	productNumber := 0
	for i, _ := range indexes {
		factor, err := numberOfProducts(lengths[:i])
		if err != nil {
			panic(err)
		}

		productNumber += factor * indexes[i]
	}

	return productNumber, nil
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