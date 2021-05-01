# Full Config File

```
# Should I be a HTTP Gateway?
GatewayEnabled: true

# Config of the HTTP Gateway
Gateway:
  Server:
    Port: 8085
    # If you want to disable Access tokens,
    # delete this entire section, otherwise, must be set in headers
    # to access any files
    # name for doc purposes only, gets hot-reloaded
    AccessTokens:
      - name: Token for ma friend Bobby
        token: secet123123

    # If you want to disable Upload tokens,
    # delete this entire section, otherwise, must be set in headers
    # to upload any files
    # name for doc purposes only, gets hot-reloaded
    UploadToken:
      - name: mytoken
        token: secreet123123


  Storage:
    # this config here is actually a minio server
    # but you can use S3 of course too
    # this is where data gets cached so that not every
    # request hits IPFS
    # I can theoretically run without this but that is
    # probably a very bad idea
    S3:
      Region: us-east-1
      Bucket: gatewaydata
      Secret: minioadmin
      Key: minioadmin
      Endpoint: http://localhost:9000
      DisableSSL: true

  # If you have an IPFS node running already,
  # set the endpoint here and tipfs will use it
  # for getting files and for communication with other peers
  # if you want to run a minimal node, delete the Backend:
  # section, and tipfs will run its own super-lightweight IPFS node
  Backend:
    IPFS:
      API: localhost:5001

  # If you do not want to set any Cors Headers,
  # delete this entire section
  CORS:
    AllowedDomains:
      - example.com

  # Enable Uploads y/n?
  # and set the maximum allowed size
  Uploads:
    Enabled: true
    MaxSize: 50 # MB



# If this section exists,
# we will run as "storage server"
# you then MUST run a full IPFS node
# if you want to store data yourself
# enter the IPFS API endpoint
PinManagerEnabled: true
PinManager:
  API: localhost:5001
  MaxSize: 50 # in MB


# DB is needed always
DB:
  Storm: /tmp/bolt.db


# General Settings
Log:
  Level: Trace
  Format: text # or json
  # logging to elastic is optional
  Elasticsearch: localhost:9200
  # file output is optional
  File: /tmp/tezos-ipfs.log


# Tell us about yourself <3
# This helps make the admin UI nicer
# only nodes that have you in "trusted Peers"
# will see this
Identity:
  Name: node1
  Organisation: TezosCommons
  Contact: "@johann on tezos-dev slack"
  Comment: "IPFS rocks"


# The Peers: Section of this config file gets reloaded immediately
# if a daemon is running, you do NOT need to restart anything
# if you make changes to this section, you SHOULD check the logs
# to make sure you have no error that makes this file invalid to parse
Peers:

  # If you run a Gateway, auto-cache content from this peers
  CacheFor:
    - 12D3KooWKpNTJYurmMnoVpLaMoiJTKHjYifeMm4BHMpNrgcWpRH2
  # If you run a storage server, auto-pin for these nodes
  # auto-pins can be overwritten
  PinFor:
    - 12D3KooWKpNTJYurmMnoVpLaMoiJTKHjYifeMm4BHMpNrgcWpRH2
  # trust information that comes from these peers and
  # If you run a Gateway, make sure tipfs is conencted to
  TrustedPeers:
    - 12D3KooWKpNTJYurmMnoVpLaMoiJTKHjYifeMm4BHMpNrgcWpRH2

```