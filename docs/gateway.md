# Gateway API

## Routes Overview

* GET `/ipfs/:cid`  fetch file
* POST `/upload`    upload file, no gurantees
* POST `/upload/once` upload file, wait for one storage confirmation
* POST `/upload/store_and_cache` wait for 1 storage and cache confirmation
* POST `/upload/threshold` upload with custom threshold
* GET `/network` returns peers we are connected to

## Fetch Data

Fetch Object from ipfs by their Content ID, if AccessTokens acre configured, must include the token
in the header `Token` field.

curl example:

```
# curl -H "Token:secret" http://127.0.0.1:8085/ipfs/QmWATWQ7fVPP2EFGu71UkfnqhYXDYH566qy47CnJDgvs8u
Hello World
```

## Upload Data

Similar to AccessTokens, if configured, all `/upload*` calls will verify the
existence of UploadTokens in the Header `Token` field.

Only the `/upload` call returns immediately, without waiting for other nodes.

curl example:

```
# curl -F "file=@./testfile" -H "Token:upload123" http://127.0.0.1:8085/upload
{"Cid":"bafybeibdm7sdv4javmutsm3fes62epzgha24rwbyqq44onqz2crlecvywm"}
```


### Upload with feedback

Calls except the simple '/upload' will wait for other nodes to confirm that they have either cached or stored
our data, the Response for all these calls is the same and looks lke this:
```
{
    "Cid": "bafybeibdm7sdv4javmutsm3fes62epzgha24rwbyqq44onqz2crlecvywm",
    "CacheNodes": [
        {
            "Name": "copycat",
            "PeerId": "12D3KooWK4Typh8no29AF2iCJ5QEDRZAbFrWLavE5v3G1NWQbKwM"
            "Cached": true # this node has cached our data
        }
    ],
    "StorageNodes": [
        {
            "Name": "storage-only",
            "PeerId": "12D3KooWSb2MqWGib529J5vR2nW9FpuNh6L2y71M9zNLwiFPa4ez",
            "Stored": true # this node has stored our data
        }
    ],
    "NumberCaches": 1, # number of cache nodes available
    "NumberStores": 1, # number of storage nodes available
    "NumberCached": 1, # how many nodes have cached our content
    "NumberStored": 1, # how many nodes have stored our content
    "Status": "Success" # or Timoeut
}

```

The timout to wait for other nodes is configured per default to be 30 seconds, and can be increased
by including the `timeout` field in the request, it is advisable to do so when uploading very large files

curl example with custom timeout of 120 seconds:

```
curl -F "file=@./verybigfile" -F "timeout=120" -H "Token:secret123" http://127.0.0.1:8085/upload/once
```

It is possible to configure custom thresholds, in this case a client might want to get the '/network' call
to first see how many nodes are available, and then set thresholds appropriately, we also provide 2 convenience routes:

####  `/upload/once`

Will wait until our content is stored on minimum one node

#### `/upload/store_and_cache`

Will wait until our content is stored on one node, and cached on one node.
( these can be different nodes )


#### `/upload/threshold`

Allows you to specify custom thresholds via POST-Form variables:

* cache
* store

curl example:

```
curl -F "file=@./testfile10" -F "cache=0" -F "store=2"  -H "Token:secret123" http://127.0.0.1:8085/upload/threshold

# with optional custom timeout
curl -F "file=@./verybigfile"  -F "cache=0" -F "store=2" -F "timeout=180" -H "Token:secret123" http://127.0.0.1:8085/upload/threshold
```


## Network

This call returns what nodes we are aware of, that either store or cache content for us.
Response is a subset of the `/upload/*` responses.

```
curl  http://127.0.0.1:8085/network | jq
{
    "CacheNodes": [
        {
            "Name": "copycat",
            "PeerId": "12D3KooWK4Typh8no29AF2iCJ5QEDRZAbFrWLavE5v3G1NWQbKwM"
        }
    ],
    "StorageNodes": [
        {
            "Name": "storage-only",
            "PeerId": "12D3KooWSb2MqWGib529J5vR2nW9FpuNh6L2y71M9zNLwiFPa4ez"
        }
    ],
    "NumberCaches": 1,
    "NumberStores": 1
}
```
