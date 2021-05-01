package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"go.uber.org/dig"
	"gopkg.in/yaml.v2"
)

func GetConfigCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "config",
	}
	root.AddCommand(GetConfigShowCommand(c),GetPublicKeyCommand(c),GetPrivateKeyCommand(c))
	return root
}

func GetConfigShowCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "show",
		Run: func(cmd *cobra.Command, args []string) {
			c.Invoke(func(c *config.Config) {
				yb,_ := yaml.Marshal(c)
				fmt.Println(string(yb))
			})
		},
	}
	return root
}

func GetPublicKeyCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "pubkey",
	}
	root.AddCommand(GetPublicKeyShowCommand(c),GetPublicIdentityShowCommand(c))
	return root
}


func GetPublicKeyShowCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "peerId",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.Invoke(func(privkey []byte, log *logrus.Entry) {
				priv,err := crypto.UnmarshalPrivateKey(privkey)
				if err != nil {
					log.Fatal(err,"1")
				}
				std,_ := crypto.PrivKeyToStdKey(priv)
				_,pub,err := crypto.KeyPairFromStdKey(std)
				if err != nil {
					log.Fatal(err)
				}
				identity,_ := peer.IDFromPublicKey(pub)
				fmt.Println("\nResult:")
				fmt.Println(identity.String())
				fmt.Println("\nPlease note:\nIf you connect this instance of tipfs to a external IPFS node, we will se its keys instead!")
			})
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	return root
}


func GetPublicIdentityShowCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "show",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.Invoke(func(privkey []byte, log *logrus.Entry) {
				priv,err := crypto.UnmarshalPrivateKey(privkey)
				if err != nil {
					log.Fatal(err,"1")
				}
				std,_ := crypto.PrivKeyToStdKey(priv)
				_,pub,err := crypto.KeyPairFromStdKey(std)
				if err != nil {
					log.Fatal(err)
				}
				pubBytes,err := crypto.MarshalPublicKey(pub)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("\nResult:")
				fmt.Println(base64.StdEncoding.EncodeToString(pubBytes))
				fmt.Println("\nPlease note:\nIf you connect this instance of tipfs to a external IPFS node, we will se its keys instead!")
			})
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	return root
}

func GetPrivateKeyCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "privkey",
	}
	root.AddCommand(GetPrivateKeyShowCommand(c))
	return root
}


func GetPrivateKeyShowCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "show",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.Invoke(func(privkey []byte, log *logrus.Entry) {
				priv,err := crypto.UnmarshalPrivateKey(privkey)
				if err != nil {
					log.Fatal(err,"1")
				}
				privBytes,err := crypto.MarshalPrivateKey(priv)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("\nResult:")
				fmt.Println(base64.StdEncoding.EncodeToString(privBytes))
				fmt.Println("\nPlease note:\nIf you connect this instance of tipfs to a external IPFS node, we will se its keys instead!")
			})
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	return root
}