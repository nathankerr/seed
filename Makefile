.PHONY: all
all:
	# go fmt
	go build -o seed
	./seed

.PHONY: clean
clean:
	-rm seed