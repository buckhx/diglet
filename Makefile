BINARY=diglet
VERSION=`git describe --always --tags`
BUILD_TIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-extldflags '-static'"
PACKAGES=`go list ./... | grep -v /vendor/`

#TODO make sure this export works correctly
export GO15VENDOREXPERIMENT := 1

.PHONY: build doc fmt lint run test clean get vet vendor vendor_clean vendor_list version list

# Default will usually be build
default: test

# Use this to ignore the vendor folder
list: 
	@echo $(PACKAGES)

build:	
	# go build -ldflags "-X 'main.version=`git describe --always` --tags' -extldflags '-static'" -o fence
	mkdir -p ./dist
	go build -v ${LDFLAGS} -o ./dist/$(BINARY)
	chmod +x ./dist/$(BINARY)

doc:
	godoc -http=:6060 -index

fmt:
	go fmt $(PACKAGES)

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint $(PACKAGES)

get:
	go get -v ./...

vet:
	go vet $(PACKAGES)

run:	
	# not sure if this works
	$(build)
	./dist/$(BINARY)

test:
	go test -v $(PACKAGES)

clean:
	rm -rf ./dist

# get and vendor update
# go get -u github.com/kardianos/govendor
vendor:
	govendor init
	govendor add +external
	govendor update +external

vendor_clean:
	rm -dRf ./vendor/*

vendor_list:
	go list ./... | grep /vendor/

version:
	@echo ${VERSION}

xcompile:
	go get -u github.com/mitchellh/gox
	mkdir -p ./dist
	gox ${LDFLAGS} -output "./dist/$(BINARY)_{{.OS}}_{{.Arch}}"
