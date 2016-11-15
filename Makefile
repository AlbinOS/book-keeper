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

docker:
	GOOS=linux GOARCH=amd64 go build -a -o build/book-keeper github.com/AlbinOS/book-keeper && cp -r views build/views && cp -r static build/static
	cd build && docker build -t albinos/book-keeper . && cd ..
	rm -rf build/book-keeper && rm -rf build/views && rm -rf build/static

docker_push:
	docker push albinos/book-keeper

docker_all: docker docker_push
