# tezos-ipfs (WIP)

`tipfs` helps coordinate storage replication, and quick distribution of objects on IPFS,
it can run with either an external, full IPFS node, or with a minimal embedded IPFS node depending on the use case.
To communicate with other instances, we only need the default IPFS PubSub implementation, if you run without a
external node this is enabled automatically, if you use an existing IPFS node, make sure it has [PubSub enabled](https://docs.ipfs.io/how-to/configure-node/#pubsub),
specifically, the newer [gossipsub](https://github.com/libp2p/specs/tree/master/pubsub/gossipsub) implementation.

# Ipfs Basics
All communication and coordination between tipfs instances happens via the IPFS network, It is a good idea to take a few minutes
to understand the [basics of IPFS](./docs/ipfs_basics.md).

# Configuration
For documentation examples, please see [Common use cases](./docs/common_use.md)

# API Docs

[Gateway API](./docs/gateway.md)


[Admin API](./docs/admin.md)

## Deployment

we offer Docker images on: https://hub.docker.com/repository/docker/tezoscommons/tezos-ipfs

Here is a simple docker-compose example with

* tezos-ipfs
* a real ipfs node ( optional )
* a minio storage server ( optional )

```
version: "3.3"
services:
  tezos-ipfs:
    depends_on:
      - minio
      - ipfs
    image: tezoscommons/tezos-ipfs:<tag>
    ports:
      - 80:80 // public gateway
    command: [ "run" ]
    volumes:
      - /path/to/config.yml:/root/config.yml
      - /data:/data # db for pins etc

  ipfs:
    image: ipfs/go-ipfs:latest
    command: ["daemon","--routing","dht","--enable-pubsub-experiment"]
    ports:
      - 4001:4001

  minio:
    image: minio/minio
    command: [ "server","/data"]
    volumes:
     - minio_data:/data
```