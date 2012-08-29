package main

import (
	"fmt"
	"log"
	// "os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
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

type transformationFn func(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection

type transformationFns map[string]transformationFn

func applyTranformations(transformations transformationFns, seeds seedCollection) budCollection {
	buds := make(budCollection)

	for sname, seed := range(seeds) {
		clusters := getClusters(sname, seed)

		for name, cluster := range(clusters) {
			transformation, ok := transformations[cluster.typ()]
			if !ok {
				fmt.Println("Tranformation for", name, cluster.typ(), "not supported!")
				// os.Exit(1)
				continue
			}

			buds = transformation(buds, cluster, seed, sname)
		}
	}

	return buds
}

var serverTranformationFns = map[string]transformationFn{
	"101": serverTransformation101,
}

func serverTransformation101(buds budCollection, cluster *cluster, seed *seed, sname string) budCollection {
	transformationinfo()

	sname = strings.Title(sname) + "Server"

	bud, ok := buds[sname]
	if !ok {
		bud = newBud()
	}

	for name, _ := range(cluster.collections) {
		collection := seed.collections[name]
		switch collection.typ {
		case seedInput:
			// replace the inputs with channels and scratches
			// this removes the need to rewrite the rules
			input := seedTableToBudTable(name, budScratch, collection)
			bud.collections[name] = input

			cname := name + "_channel"
			channel := seedTableToBudTable(cname, budChannel, collection)
			bud.collections[cname] = channel

			rewrite := newRule(collection.source)
			rewrite.value = fmt.Sprintf("%s <= %s.payloads", name, cname)
			bud.rules = append(bud.rules, rewrite)
		case seedOutput:
			// replace the outputs with channels and scratches
			output := seedTableToBudTable(name, budScratch, collection)
			bud.collections[name] = output

			cname := name + "_channel"
			channel := seedTableToBudTable(cname, budChannel, collection)
			bud.collections[name] = channel

			rewrite := newRule(collection.source)
			rewrite.value = fmt.Sprintf("%s <~ %s.payloads", cname, name)
			bud.rules = append(bud.rules, rewrite)
		case seedTable:
			table := seedTableToBudTable(name, budPersistant, collection)
			bud.collections[name] = table
		}

		buds[sname] = bud
	}

	return buds
}