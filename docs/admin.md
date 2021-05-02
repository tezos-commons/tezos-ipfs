# Admin API

Using the default config, the Admin API runs on `http://localhost:5082` and should not be exposed publicly.

## Routes

* POST `/pin/:cid` create pin
* DELETE `/pin/:cid` delete pin
* POST `/pin/:cid/block` block content
* GET `/id` get peerID

### Create Pin

POST `/pin/:cid` will pin the cid provided and also broadcast a pin request to others

### Delete Pin

DELETE `/pin/:cid` will delete the pin only locally, if you are running a gateway, this cid can still be fetched

### Block Pin

POST `/pin/:cid/block` will delete content locally, and prevent to gateway from serving this data to others,
if this cid is requested, it will return a 404 response

### Peer ID

GET `/id`  returns the local peerId as base58 encoded string
