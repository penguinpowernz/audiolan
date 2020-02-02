build:
	go build -o bin/audiolan ./cmd/audiolan
	go build -o bin/audiolan-cli ./cmd/audiolan-cli

reqs:
	sudo apt-get install xserver-xorg-input-libinput-dev xserver-xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev pkg-config libgtk-3-dev libgtk2.0-dev libgtkd-3-dev libgtkgl2.0-dev libgl1-mesa-dev portaudio19-dev

dist:
	go build -o bin/audiolan.Linux.amd64 ./cmd/audiolan
	go build -o bin/audiolan-cli.Linux.amd64 ./cmd/audiolan-cli
	# GOARCH=arm go build -o bin/audiolan.Linux.armhf ./cmd/audiolan
	# GOARCH=arm go build -o bin/audiolan-cli.Linux.armhf ./cmd/audiolan-cli
	# GOARCH=386 go build -o bin/audiolan.Linux.i386 ./cmd/audiolan
	# GOARCH=386 go build -o bin/audiolan-cli.Linux.i386 ./cmd/audiolan-cli