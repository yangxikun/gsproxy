gsproxy is a simple http proxy with basic authentication support.

Installing from source
----------------------

To install, run

    $ go get github.com/yangxikun/gsproxy

Build

    $ go install github.com/yangxikun/gsproxy/cmd/gsproxy 

You will now find a `gsproxy` binary in your `$GOPATH/bin` directory.

Usage
-----

Start proxy

    $ gsproxy -auth test:1234

Run `gsproxy -help` for more information.