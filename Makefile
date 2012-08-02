.PHONY: all
all:
	# go fmt
	go build -o seed
	-rm -rf bud
	./seed kvs.seed
	cat bud/kvsserver.rb

.PHONY: clean
clean:
	-rm seed
	-rm -rf bud