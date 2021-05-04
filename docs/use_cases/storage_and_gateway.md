# Storage and Gateway Example

In this example, we want to run a Storage Server that can serve content to other IPFS nodes as well as a HTTP Gateway, for this you need
to run a real IPFS node and specify the API endpoint if it in the config file ( default port: 5001 ).


## Config file

```

GatewayEnabled: true
PinManagerEnabled: true
PinManager:
  API: localhost:5001 # this is the API port of the IPFS node

Gateway:
  Server:
    Port: 8085
  Storage: # Where we should cache files we got from IPFS, optional since we have a storage node anyways
    S3:
      Region: us-east-1
      Bucket: gatewaydata
      Secret: minioadmin
      Key: minioadmin
      Endpoint: http://localhost:9000 # AWS region endpoint or minio port
      DisableSSL: true
  Uploads:
    Enabled: true
    
      
Peers:
  # Here, make a list of all the peers you want to store data for
  PinFor:
    - 12D3K....
    
  # Here, make a list of all the peers you want to cache data for ( on S3 )
  # only necessary when using the Gateway and you want to use S3 cache
  CacheFor:
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

for more configuration options like storage and access tokens plese take a look at the [full config file](./full_config.md)