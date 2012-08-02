.PHONY: all
all: seed

print: *.go
	./gen-print.sh
	context print
	open print.pdf

seed: *.go
	# go fmt
	go build -o seed
	-rm -rf bud
	./seed kvs.seed
	cat bud/kvsserver.rb

.PHONY: clean
clean:
	-rm seed
	-rm -rf bud
	-rm -rf tmp
	-mkdir tmp
	context --purge
	-rm print.tex print.pdf 
