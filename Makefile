build:
	tinygo build -target wasi -o ocigetterplugin.wasm main.go
	#GOOS=wasip1 GOARCH=wasm go build -o . ./...

test:
	extism call ocigetterplugin.wasm greet --memory-max 65536 --input "$$(cat test_input.json)" --wasi
