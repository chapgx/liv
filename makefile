





build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux_amd64 -ldflags="-s -w" . && \
	GOOS=linux GOARCH=arm64 go build -o ./bin/linux_arm64 -ldflags="-s -w" .
