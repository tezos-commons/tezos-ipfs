# API

## Gateway


The gateway exposes the following Routes:

* `GET /ipfs/:cid` Get a IPFS object
* `GET /network` Get information about the tipfs network, other peers
* `POST /upload` Simple Upload, no guarantees


## Curl examples

```
# uplaoding /testfile with UploadToken
curl -v -F "file=@/testfile" -H "Token:secret123" http://127.0.0.1:8085/upload

# without token
curl -v -F "file=@/testfile" http://127.0.0.1:8085/upload
```



