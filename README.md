# Brot-Client

![Build](https://github.com/TeamBrot/client/actions/workflows/go.yml/badge.svg)

![](brot-icup.jpg)

This repository contains our client for `spe_ed`, the game of the [InformatiCup 2021](https://github.com/InformatiCup/InformatiCup2021).

## Installation

This program uses the websocket library that can be found here: [https://github.com/gorilla/websocket](https://github.com/gorilla/websocket).

To install the library, run `go get github.com/gorilla/websocket`. 

## Build & Run

The client code is located in the `client/` directory. Before building, go there by running `cd client`.

To build the code, run `go build`. Afterwards, you can run the client with `./client`.

## Playing locally

### Server

For development, you will also need a `spe_ed` server. We built a development server that can be found [here](https://github.com/TeamBrot/server).

The client's default server url is `ws://localhost:8080`. If you are using any other URL, please use the `URL` environment variable.

### Starting the client

You can start the client by first going into the `client` directory and then running `./client`.

This runs the `combi` client. Other clients can be run with `./client -client <client>`.

The following clients are available:

- `basic`
- `minimax`
- `rollouts`
- `probability`
- `combi`

## Playing on the official server

To play on the official server, the environment variables `URL`, `TIME_URL` and `KEY` have to be set to the appropriate values:

`URL="wss://msoll.de/spe_ed" TIME_URL="https://msoll.de/spe_ed_time" KEY="<key>" ./client`

## Docker

To build the docker image, run `docker build . -t spe_ed`.

To run the client container, run `docker run -e URL="wss://msoll.de/spe_ed" -e TIME_URL="https://msoll.de/spe_ed_time" -e KEY="<key>" spe_ed`

## Extensions

### Testing against other clients

Run `./test_internal.sh` to start mutltiple games. This can be useful to test different parameters or to run statistical analytics.

Just change the script for your purposes.

### Testing on the official API

You can also run `./test_api.sh` to play multiple games without having to restart the client manually.

The default client is the `combi` client. You can change that by setting the variable `client` directly in the script.

