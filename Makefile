# Makefile for building and testing this provider.
#
# "build" target builds an unversioned provider for the current
# OS and architecture with flags appropriate for debugging.
#
# "install" installs the unversioned provider in the local
# Terraform cache directory so it can be found by "terraform init".
#
# "test" and "testacc" do what you expect.
#
# "release" uses goreleaser to release the semver tag that the
# current branch corresponds to (make sure to push up the tag first).

TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=localhost.arpio.io
NAMESPACE=arpio
NAME=arpio
BINARY=terraform-provider-${NAME}
VERSION=0.0.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

default: install

clean:
	go clean

build:
	go build -gcflags="all=-N -l" -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	go test $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   

release:
	goreleaser release --rm-dist
