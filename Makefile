install:
	go install -ldflags '-s -w' ./cmd/message-hub

install-debug:
	go install -race ./cmd/message-hub
