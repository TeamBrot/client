# Brot-Client

![Build](https://github.com/TeamBrot/client/actions/workflows/go.yml/badge.svg)
![LOC](https://img.shields.io/tokei/lines/github/TeamBrot/client)
![Go Version](https://img.shields.io/github/go-mod/go-version/TeamBrot/client)
![License](https://img.shields.io/github/license/TeamBrot/client)

[![](brot-icup.png)](https://teambrot.github.io/)

This repository contains our client for `spe_ed`, the game of the [InformatiCup 2021](https://github.com/InformatiCup/InformatiCup2021).

## Overview
In this README you find all the information you need to use our software. Navigate around with the table of contents. For more detailed information read the wiki or follow the links in the README. If you think some important information or feature is missing, feel free to open an [issue](https://github.com/TeamBrot/client/issues). If you'd like to give us feedback directly feel free to mail us.
- [Manual Installation](#installation)
- [Docker Installation](#docker)
- [Extensions](#extensions)

## Installation

After completing this, you set up a `spe_ed` development server and you can run our client locally.

### 0. Prerequisites

Our software is written in Go. So you need an up-to-date go installation. You can get information on how to install Go on your computer [here](https://golang.org/doc/install). At least version 1.15 is required.

This program uses the websocket library that can be found here: [https://github.com/gorilla/websocket](https://github.com/gorilla/websocket).

After installing Go, you can run `go get github.com/gorilla/websocket` to install the library.

### 1. Cloning the repository

Run `git clone https://github.com/TeamBrot/client.git` to clone the repository. 

Run `cd client` to jump right into it.

### 2. Build

The client code is located in the `client` directory. Before building, go there by running `cd client`.

To build the code, run `go build`. In case of failure, check if you are in the right directory because the repository is also named client. If you change the code of the client make sure to run the build command.

### 3. Setting up the development server

To run the client you need a server, that implements the game API. If you got your own `spe_ed` server you can skip this step. Otherwise, you can use ours which we wrote for development purposes. You can find it [here](https://github.com/TeamBrot/server). 

### 4. Running the Client

You can start the client by going into the `client` directory and then running `./client`. Now the client tries to connect to our development server. If you get a connection established message you've succesfully setup our software 🥳

If you wish to connect to another server, set the `URL` environment variable accordingly. You find information about that [here](#connecting-to-other-servers)

`./client` runs the `combi` client. Other clients can be run with `./client -client <client>`.

The following clients are available:

- [`basic`](https://github.com/TeamBrot/client/wiki/basic)
- [`minimax`](https://github.com/TeamBrot/client/wiki/minimax)
- [`rollouts`](https://github.com/TeamBrot/client/wiki/rollouts)
- [`probability`](https://github.com/TeamBrot/client/wiki/probability-tables)
- [`combi`](https://github.com/TeamBrot/client/wiki/combi)

### 5. Connecting to other servers

To play on other servers, the environment variables `URL`, `TIME_URL` and `KEY` have to be set to the appropriate values. 

For the official API of the competition, use these values:

`URL="wss://msoll.de/spe_ed" TIME_URL="https://msoll.de/spe_ed_time" KEY="<key>"`

If you connect the client to another server, you can watch the game in your browser under [http://localhost:8081/](http://localhost:8081)

## Docker

To build the docker image, run `docker build . -t spe_ed`.

To run the client container, run `docker run -e URL="wss://msoll.de/spe_ed" -e TIME_URL="https://msoll.de/spe_ed_time" -e KEY="<key>" spe_ed`.

## Extensions

We built several extensions for our `spe_ed` client. For example several scripts that can run or visualize games.

### Testing against other clients

Run `./test_internal.sh` to start mutltiple games. This can be useful to test different parameters or to run statistical analytics.

You can stop the script by creating a file with the name `stop` at the root of the repository.

Just change the script for your purposes.

### Testing on the official API

You can also run `./test_api.sh` to play multiple games on the official without having to restart the client manually.

The default client is the `combi` client. You can change that by setting the variable `client` directly in the script.

### Visualize games as videos

Prerequisites:
- Python 3.9+
- Pillow
- ffmpeg

Run `./visualize.py <path to JSON-Log>` to create a video from a JSON-Log. You typically find all log files under `client/log`. To see all command line options run `./visualize.py --help`.
