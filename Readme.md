# tezos-ipfs (WIP)

`tipfs` helps coordinate storage replication, and quick distribution of objects on IPFS,
it can run with either an external, full IPFS node, or with a minimal embedded IPFS node depending on the use case.
To communicate with other instances, we only need the default IPFS PubSub implementation, if you run wihtout a
external node this is enabled automatically, if you use an existing IPFS node, make sure it has [PubSub enabled](https://docs.ipfs.io/how-to/configure-node/#pubsub),
specifically, the newer [gossipsub](https://github.com/libp2p/specs/tree/master/pubsub/gossipsub) implementation.

# Ipfs Basics
All communication and coordination between tipfs instances happens via the IPFS network, It is a good idea to take a few minutes
to understand the [basics of IPFS](./docs/ipfs_basics.md).

# Configuration
For documentation examples, please see [Common use cases](./docs/common_use.md)

# API Docs

[Gateway API](./docs/gateway.md)