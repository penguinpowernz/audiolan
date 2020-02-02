# Audiolan

A server and client to deliver audio playing the servers speakers to the clients speakers via websockets. Some
use cases could be:

* distributing audio to multiple slave Raspberry Pis
* play audio from a PC with no speakers attached

Binaries can be found in the releases section.

## Build

To build, first need to install the Debian packages:

    make reqs

Then build:

    make build

## Usage

To run the app the following package needs to be installed:

```
sudo apt-get install portaudio19-dev
```

### CLI

Using the CLI version you can run the server like so on the computer whose default sound output (read: speakers) you want to listen to remotely

```
audiolan-cli -s
```

On the computer you want to listen from run this command:

```
audiolan-cli -c 192.168.1.100:4567
```

This will being playing any audio from the first computer on the second.

### UI

When running the UI verison on the computer whose sound you want to send out, 

![](https://raw.githubusercontent.com/penguinpowernz/audiolan/master/server.png)

1. choose the **Server** tab
2. enter the port number to run on (optional)
3. click **Start** to start listening for requests

When running the UI version on the computer from which you want to listen:

![](https://raw.githubusercontent.com/penguinpowernz/audiolan/master/client.png)

1. choose the **Client** tab
2. enter the IP and port number to listen to (e.g. 192.168.1.100:4567)
3. click the **Connect** button to start listening to audio

## TODO

- [x] fix stuttering/buffering issue
- [x] serve audio to single client
- [ ] show multiple clients in the UI
- [ ] make log available from the UI
- [ ] authentication
- [ ] test on Windows
- [ ] test on MacOS
