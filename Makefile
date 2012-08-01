.PHONY: all
all:
	# go fmt
	go build -o seed
	./seed kvs.seed

.PHONY: clean
clean:
	-rm seed