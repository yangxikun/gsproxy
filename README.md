gsproxy is a simple http proxy with basic authentication support.

Installing from source
----------------------

To install, run

    $ go install github.com/yangxikun/gsproxy/cmd/gsproxy@latest 

You will now find a `gsproxy` binary in your `$GOPATH/bin` directory.

Usage
-----

Start proxy

    $ gsproxy --auth test1:1234,test2:5678

Run `gsproxy -help` for more information:

    Usage of gsproxy:
      --credentials string   basic credentials: username1:password1,username2:password2
      --gen_credential       generate a credential for auth
      --listen string        proxy listen addr (default ":8080")

Config by environment variable:

    GSPROXY_LISTEN=:9898
    GSPROXY_CREDENTIALS=test1:1234,test2:5678
    GSPROXY_GEN_CREDENTIAL=true

Config by yaml file:

    listen: :8181
    credentials:
      - aaa:bbb
      - ccc:ddd
    gen_credential: false