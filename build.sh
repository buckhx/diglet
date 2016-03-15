#! /bin/bash

repeat=$(printf "%40s")
function header {
    echo "${repeat// /*} $1 ${repeat// /*}"
}

header "go generate"
go generate

header "go fmt"
gofmt -s -w -d ./

header "go get"
go get github.com/fzipp/gocyclo
go get github.com/client9/misspell/cmd/misspell
go get github.com/golang/lint/golint
go get ./...

header "style checks"
header "misspell"
echo "misspell ./**/*"
misspell ./**/*

header "go lint"
echo golint ./...
golint ./...

header "go vet"
echo go tool vet ./
go tool vet ./

header "go cyclo"
echo gocyclo -over 10 ./
gocyclo -over 10 ./

header " go test"
echo go test ./...
go test ./...

header "go build"
echo go build -ldflags '-extldflags "-static"' -o diglet
go build -ldflags '-extldflags "-static"' -o diglet

echo "Done!"
