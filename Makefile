dependencies:
	go get -u -f ./...

clean:
	rm -f $(GOPATH)/bin/book-keeper

test:
	go test ./...

.PHONY: build
build:
	go build -a -o $(GOPATH)/bin/book-keeper github.com/AlbinOS/book-keeper

install: dependencies build
