.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/utchain/go-utchain/cmd/swarm
	go build -o bin/gut     github.com/utchain/go-utchain/cmd/gut
	go build -o bin/bootnode github.com/utchain/go-utchain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/utchain/go-utchain/swarm/...
