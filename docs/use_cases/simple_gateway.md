# Gateway Configuration

A Gateway connects to the IPFS Network and serves Data from it via an HTTP API,

## Config

```
GatewayEnabled: true
Gateway:
  Server:
    Port: 8085
  Storage: # Where we should cache files we got from IPFS  
    S3:
      Region: us-east-1
      Bucket: gatewaydata
      Secret: minioadmin
      Key: minioadmin
      Endpoint: http://localhost:9000 # AWS region endpoint or minio port
      DisableSSL: true
  Uploads:
    Enabled: true

DB:
  Storm: /data/bolt.db

Identity:
  Name: gw1
  Organisation: TezosCommons
  Comment: "IPFS rocks"

Peers:
  # The peerIds of Nodes that we want to automatically cache
  # content for
  CacheFor:
    - 12D3KooWKpNTJYurmMnoVpLaMoiJTKHjYifeMm4BHMpNrgcWpRH2

```