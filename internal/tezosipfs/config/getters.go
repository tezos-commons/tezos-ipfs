package config

import (
	"os"
)

func (c *Config) GetIpfsAPI() *string {
	var res *string
	if c.PinManagerEnabled && c.PinManager.API != "" {
		res = &c.PinManager.API
		return res
	}
	if c.GatewayEnabled && c.Gateway.Backend.IPFS.API != "" {
		res = &c.Gateway.Backend.IPFS.API
		return res
	}
	if c.GatewayEnabled && c.Gateway.Backend.IPFS.API != "" {
		if c.PinManagerEnabled && c.PinManager.API != "" {
			if c.Gateway.Backend.IPFS.API != c.PinManager.API {
				c.log.Fatal("Gateway and Pin Manager can not be different IPFS Nodes!")
				os.Exit(1)
			}
		}
	}

	// can return nil, in which case
	// run our own libp2p instance
	return res
}
