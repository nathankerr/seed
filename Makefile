# NOTE: this makefile is only used for building seed.leg.go

all: seed.leg.go
	rm -f ${GOPATH}/bin/seed
	rm -rf build
	go install ./...
	go fmt ./...
	go vet ./...
	seed -t "go seed fieldgraph graph" -transformations "networkg" services/kvs/kvs.seed
	cd build; dot -T pdf -O kvsserver.fieldgraph.dot; open kvsserver.fieldgraph.dot.pdf

seed.leg.go: seed.leg
	leg $< > $@

test:
	go test