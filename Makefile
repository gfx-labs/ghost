

build:
	go build -o spook.out ./cmd/spook

install:
	mv spook.out $(HOME)/.local/bin/spook
