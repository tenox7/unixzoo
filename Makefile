all: unixzoo

unixzoo: *.go index.tmpl go.mod go.sum
	go build .

test:
	go test ./...

cross:
	GOOS=linux GOARCH=amd64 go build -o unixzoo-amd64-linux .
	GOOS=linux GOARCH=arm64 go build -o unixzoo-arm64-linux .
	GOOS=darwin GOARCH=arm64 go build -o unixzoo-arm64-macos .
	GOOS=freebsd GOARCH=amd64 go build -o unixzoo-amd64-freebsd .

clean:
	rm -f unixzoo unixzoo-*
