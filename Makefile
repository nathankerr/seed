# NOTE: this makefile is only used for building seed.leg.go

all: seed.leg.go
	rm -f /Users/alaster/Projects/seed/bin/seed
	rm -rf build
	go install ./...
	go fmt ./...
	seed -t "go seed" -transformations "" examples/kvs/kvs.seed

seed.leg.go: seed.leg
	leg $< > $@