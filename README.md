test-cec.go
===========

Simple demo code for using & testing out `github.com/chbmuc/libcec`.

> [!NOTE]  Outdated / unmaintained dependencies!
> `github.com/chbmuc/libcec` only supports `libcec` version `4.x`!
> As such, this project will not work on most modern Linux distros.
> The instructions were originally tested for Ubuntu `16.04`, Go `1.7.4`

Building
========

First, make sure you have [`golang` installed][install-golang]


Next, install dependencies:

    sudo apt-get -y install build-essential git libcec-dev libp8-platform-dev libudev-dev

Optionally, you may wish to install `cec-client` for comparing how it interacts with CEC:

    sudo apt-get -y install cec-utils

Make sure you have [`GOPATH` set up][setup-gopath] somewhere where you would like to [organize your go code][how-to-write-go]. Then:

    go get github.com/chbmuc/cec
    go get github.com/trinitronx/test-cec
    cd $(go env GOPATH)/src/github.com/trinitronx/test-cec
    go install

Running
=======

It's as simple as:

    $(go env GOPATH)/bin/test-cec


[install-golang]: https://github.com/golang/go/wiki/Ubuntu
[setup-gopath]: https://github.com/golang/go/wiki/SettingGOPATH
[how-to-write-go]: https://golang.org/doc/code.html
