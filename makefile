build-all:
	go build -o sql_engine_linux .
	GOOS=windows GOARCH=amd64 go build -o sql_engine_windows.exe .
	GOOS=darwin GOARCH=amd64 go build -o sql_engine_mac .
