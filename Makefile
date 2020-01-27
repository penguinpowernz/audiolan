build:
	go build -o bin/audiolan ./cmd/audiolan

reqs:
	sudo apt-get install xserver-xorg-input-libinput-dev xserver-xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev pkg-config libgtk-3-dev libgtk2.0-dev libgtkd-3-dev libgtkgl2.0-dev libgl1-mesa-dev portaudio19-dev

