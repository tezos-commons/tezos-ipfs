# Upload to anoter node

Here we assume that you want to host only a minimal gateway and allow users
to upload and download data, but you have an agreement with someone else to store data for you.


## Gateway config

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

  # If a peer is in this list,
  # or in the trusted peers of one of the peers here configured ( 2nd level )
  # we will trust data from and this node,
  # and also allow them to store data for us
  TrustedPeers:
    - <StorageNode>
    
```

After you start your Node for the first time the logs will show your PeerID,
give this PeerId to the party that stores data for you, after they add it to their
`PinFor` Section their node will appear in the '/network' API call and whenever you upload data,
this node will be notified. For more information please check the [Gateway API docs.](./../gateway.md)