# How networking works in tipfs

Depending on the use-case, you may want to run `tipfs` with or without a "real" IPFS node,
for example if you want to store data on the IPFS network, lets say you have a very large DAPP that needs
a lot of storage, you MUST run your own IPFS node, and tell `tipfs` to connect to it, if however, you run
an app where you only need to fetch data, we need to connect to the IPFS network somehow, but we do not need a "real"
IPFS node to do so, in this case, `tipfs` would run in ligt-client-mode

Generally speaking, if you want to contribute to the IPFS network and serve data to other peers on the
network, you need to run in remote-IPFS mode, everything else can be light-client-mode or remote-IPFS-mode

## `tipfs` networking modes

### remote-IPFS mode

This mode is enabled once you specify and use a IPFS node in the config file, **in this mode
`tipfs` will "reuse" the networking stack of your IPFS node, including its peer-ID.** This means 
that to other peers in the network, real nodes as well as `tipfs` instances, it will be acessible via 
the peer-ID of the IPFS node.


### light-client-mode

If `tipfs` can not find any configuration for a remote node in its config, it will setup a very very
minimal libp2p host, that implements just barely enough to communicate with the rest of the network, specifically:

* a KAD DHT
* the pub sub protocol
* the bitswap protocol

A common use case is a caching HTTP Gateway that connects to well-known nodes in the tezos-ecosystem, and lets you
fetch IPFS objects via HTTP. In light-client-mode such a Gateway could also listen to specific `ipfs` nodes, and once they
add a new object to IPFS automatically cache it, before you have even asked for it, but more on that in the config Section.
Since here we do not serve data to IPFS, we do not need a full node, we can use the much more efficient light-client-mode instead.

In this mode `tipfs` will generate its own peer ID, you can get information about this id via the CLI:

```
$ tipfs config pubkey show

Result:
CAESIOQHbGTaGHQmtRRjUakp3614C9/VXhDtQIi2dSYt9lx5
```



