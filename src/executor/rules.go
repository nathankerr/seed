package main

import (
	"errors"
	"service"
)

// builds a new collection from the input collections according to the rule
func runRule(collections map[string]*collection, service *service.Service, rule *service.Rule) *collection {
	results := newCollection(rule.Supplies, service.Collections[rule.Supplies])

	// setup index based access to the collections and rows
	productCollections := []string{}
	productCollectionLengths := []int{}
	allRows := map[string][][]interface{}{}
	for _, collectionName := range rule.Requires() {
		collectionRows := collections[collectionName].rows

		productCollections = append(productCollections, collectionName)
		productCollectionLengths = append(productCollectionLengths, len(collectionRows))

		// assign row numbers for each row in the collection
		for _, row := range collectionRows {
			allRows[collectionName] = append(allRows[collectionName], row)
		}
	}

	// find out how many rows result from the product of the collections
	numberOfProductRows, err := numberOfProducts(productCollectionLengths)
	if err != nil {
		panic(err)
	}

	// loop through the input rows (result of the product)
	for productNumber := 0; productNumber < numberOfProductRows; productNumber++ {
		// get the indexes for each collection from the productNumber
		indexes, err := indexesFor(productNumber, productCollectionLengths)
		if err != nil {
			panic(err)
		}

		// get the rows for this product
		rows := make(map[string][]interface{}) // collectionName: row
		for index, rowNumber := range indexes {
			collectionName := productCollections[index]
			rows[collectionName] = allRows[collectionName][rowNumber]
		}

		// determine if this product should be skipped
		skip := true
		if len(rule.Predicate) == 0 {
			skip = false
		}
		for _, constraint := range rule.Predicate {
			// get the left row
			lqc := constraint.Left
			leftColumnIndex := collections[lqc.Collection].columns[lqc.Column]
			left := rows[lqc.Collection][leftColumnIndex]

			// get the right row
			rqc := constraint.Right
			rightColumnIndex := collections[rqc.Collection].columns[rqc.Column]
			right := rows[rqc.Collection][rightColumnIndex]

			if left == right {
				skip = false
			}
		}

		// add the columns to the result row if the product is not skipped
		if !skip {
			result := []interface{}{}
			for _, qc := range rule.Projection {
				columnIndex := collections[qc.Collection].columns[qc.Column]
				result = append(result, rows[qc.Collection][columnIndex])
			}
			results.addRow(result)
		}
	}

	return results
}

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
