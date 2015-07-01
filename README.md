USSS: uwsgi subscription server switcher
========================================

This small software listens for udp packets on a given port and then
forwards these to a set of hosts via udp. The set of hosts is discovered
by resolving a dns name and the data is sent to all the addresses that are
returned.

Rationale
=========

I needed something that could manage multiple [uwsgi](https://uwsgi-docs.readthedocs.org/en/latest/) workers subscribing
to a variable set of [subscription servers](http://uwsgi.readthedocs.org/en/latest/SubscriptionServer.html)
to perform load balancing for some web applications. The case for disconnecting/dying 
workers is already handled by the subscription server, but there is no way of telling
an already running worker to subscribe to another subscription server without having to
restart it, reconfigure it or in any case touch it. With this hack the subscription
server ip for the workers is always fixed to the host that is running usss, so that
when a subscription server changes the worker has no knowledge of it and transparently
registers to all available subscription server/frontends.

Requirements and installation
=============================

It's written in go so it should JustWork(tm) with go > 1.4.
Just run

    go build .

in the checked out dir to build. I decided against an installable package since this is
a binary that is standalone and will be rarely updated.

Usage
=====

A typical usage might be

    usss -dnsname=somename.com -ttl=30 -outport=6262 -inport=2626 -bindaddr=10.0.0.1

so that all packets arriving on port 2626 of 10.0.0.1 will be forwarded to port 6262
on the hosts that are pointed by the somename.com record. Since the ttl is set to 30
seconds, the somename.com record will be re-resolved every 30 seconds and the list of
addresses updated accordingly.

Guarantees
==========

The goroutines that resolve the dns record and that send the data are mutually excluded via [sync.Mutex](http://golang.org/pkg/sync/)
so there is the guarantee that the set of ips that are used as destinations will be consistent
between resolves.

Bugs
====

Must have some bugs, pull requests and issues welcome!
