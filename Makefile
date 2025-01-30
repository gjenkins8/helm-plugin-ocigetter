build:
	GOOS=wasip1 GOARCH=wasm go build -o . ./...

test:
	go test ./testdriver
