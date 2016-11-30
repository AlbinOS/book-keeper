dependencies_ui:
	go get -u github.com/AlbinOS/book-keeper-ui/dist

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
	GOOS=linux GOARCH=amd64 go build -a -o build/book-keeper github.com/AlbinOS/book-keeper && cp -r dist build/dist
	cd build && docker build -t albinos/book-keeper . && cd ..
	rm -rf build/book-keeper && rm -rf build/dist

docker_push:
	docker push albinos/book-keeper

docker_all: docker docker_push
