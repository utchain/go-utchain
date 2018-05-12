# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gut android ios gut-cross swarm evm all test clean
.PHONY: gut-linux gut-linux-386 gut-linux-amd64 gut-linux-mips64 gut-linux-mips64le
.PHONY: gut-linux-arm gut-linux-arm-5 gut-linux-arm-6 gut-linux-arm-7 gut-linux-arm64
.PHONY: gut-darwin gut-darwin-386 gut-darwin-amd64
.PHONY: gut-windows gut-windows-386 gut-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gut:
	build/env.sh go run build/ci.go install ./cmd/gut
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gut\" to launch gut."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gut.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Gtst.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gut-cross: gut-linux gut-darwin gut-windows gut-android gut-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gut-*

gut-linux: gut-linux-386 gut-linux-amd64 gut-linux-arm gut-linux-mips64 gut-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-*

gut-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gut
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep 386

gut-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gut
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep amd64

gut-linux-arm: gut-linux-arm-5 gut-linux-arm-6 gut-linux-arm-7 gut-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep arm

gut-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gut
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep arm-5

gut-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gut
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep arm-6

gut-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gut
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep arm-7

gut-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gut
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep arm64

gut-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gut
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep mips

gut-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gut
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep mipsle

gut-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gut
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep mips64

gut-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gut
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gut-linux-* | grep mips64le

gut-darwin: gut-darwin-386 gut-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gut-darwin-*

gut-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gut
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gut-darwin-* | grep 386

gut-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gut
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gut-darwin-* | grep amd64

gut-windows: gut-windows-386 gut-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gut-windows-*

gut-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gut
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gut-windows-* | grep 386

gut-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gut
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gut-windows-* | grep amd64
