.DEFAULT: build

ocigetterplugin.wasm:
	GOOS=wasip1 GOARCH=wasm go build -o . ./...
	mv ocigetterplugin ocigetterplugin.wasm

build: ocigetterplugin.wasm

test:
	go test ./testdriver
