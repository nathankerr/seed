package main

import (
	"fmt"
	"log"
	"path"
	"path/filepath"
	"runtime"
)

// toggle on and off by commenting the first return statement
func transformationinfo(args ...interface{}) {
	return
	info := ""

	pc, file, line, ok := runtime.Caller(1)
	if ok {
		basepath, err := filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		sourcepath, err := filepath.Rel(basepath, file)
		if err != nil {
			panic(err)
		}
		info += fmt.Sprintf("%s:%d: ", sourcepath, line)

		name := path.Ext(runtime.FuncForPC(pc).Name())
		info += name[1:]
		if len(args) > 0 {
			info += ": "
		}
	}
	info += fmt.Sprintln(args...)

	log.Print(info)
}

type seedToBudTransformation func(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection
type seedToBudTransformations map[string]seedToBudTransformation

func applySeedToBudTranformations(seeds seedCollection, transformationList ...seedToBudTransformations) budCollection {
	buds := make(budCollection)

	for _, transformations := range transformationList {
		for sname, seed := range seeds {
			clusters := getClusters(sname, seed)

			for name, cluster := range clusters {
				transformation, ok := transformations[cluster.typ()]
				if !ok {
					fmt.Println("Tranformation for", name, cluster.typ(), "not supported!")
					continue
				}

				buds = transformation(buds, cluster, seed, sname)
			}
		}
	}

	return buds
}
