# NOTE: this makefile is only used for building seed.leg.go

all: seed.leg.go
	rm -f ${GOPATH}/bin/seed
	rm -rf build
	go install ./...
	go fmt ./...
	go vet ./...
	seed -t "go seed" -transformations "networkg" services/kvs/kvs.seed

seed.leg.go: seed.leg
	leg $< > $@

test:
	go test