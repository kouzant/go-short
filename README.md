go-short
========

A simplistic URL shortener

To listen on privileged port:

   setcap 'cap_net_bind_service=+ep' /usr/share/go-short/go-short


Build
-----

    $ go build -o go-short