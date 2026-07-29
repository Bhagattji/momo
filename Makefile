BINARY := momo
VERSION ?= dev

build:
	go build -trimpath -ldflags "-s -w -X 'momo/internal/version.Version=$(VERSION)'" -o $(BINARY) ./cmd

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X 'momo/internal/version.Version=$(VERSION)'" -o $(BINARY)-linux-amd64 ./cmd

build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X 'momo/internal/version.Version=$(VERSION)'" -o $(BINARY)-windows-amd64.exe ./cmd

clean:
	-rm -f $(BINARY) $(BINARY)-* || del $(BINARY) || true
	go clean

test:
	go test ./...

release-local:
	goreleaser --snapshot --skip-publish --rm-dist
