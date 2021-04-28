package network

const BROADCAST_TOPIC = "TEZOS_IPFS"

type Message struct {

}

func GetNetwork(ipfsClient *IPFS, lightclient *Lightclient) NetworkInterface {
	if ipfsClient == nil {
		lightclient.Setup()
		return lightclient
	} else {
		return ipfsClient
	}
}