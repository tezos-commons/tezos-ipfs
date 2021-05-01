# Storage Example

In this example, we want to run a Storage Server that others can pin files to,
if you want to store data on IPFS you need to run a real IPFS node, and tell `tipfs` about it
via the config file. You need to know the peerId of the Nodes you want to store data for

## Config file

```
# we disable the Gateway
GatewayEnabled: false

# and enable the pinManager
PinManagerEnabled: true
PinManager:
  API: localhost:5001 # this is the API port of the IPFS node
  
Peers:
  # Here, make a list of all the peers you want to store data for
  PinFor:
    - 12D3K....
    
  # Optional, but if you know and trust other nodes
  # you can add them here, this will help with connectivity
  TrustedPeers:
    - 12D3....
    
    
## General Settings

Identity:
  Name: storage-Only
  Organisation: TezosCommons
  Comment: "IPFS rocks"

# stores your pins
DB:
  Storm: /data/storm.db

# Admin API, do not make this public!
Admin:
  Host: 127.0.0.1
  Port: 5082
```