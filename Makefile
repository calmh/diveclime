.PHONY:	build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -buildvcs=false -ldflags '-w -s' -o viva-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -v -buildvcs=false -ldflags '-w -s' -o viva-linux-arm64
