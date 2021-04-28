package cmd

import (
	"encoding/base64"
	"fmt"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

func GetToolsCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "tools",
	}
	root.AddCommand(GetGenKeysCommand(c))
	return root
}

func GetGenKeysCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "genkeys",
		Run: func(cmd *cobra.Command, args []string) {

			priv, pub, _ := crypto.GenerateKeyPair(crypto.Ed25519, 256)
			bpriv,_ := crypto.MarshalPrivateKey(priv)
			bpub,_ := crypto.MarshalPublicKey(pub)
			identity,_ := peer.IDFromPublicKey(pub)

			fmt.Println("\nIdentity:")
			fmt.Println(identity)

			fmt.Println("\nPublicKey:")
			fmt.Println(base64.StdEncoding.EncodeToString(bpub))

			fmt.Println("\nPrivateKey:")
			fmt.Println(base64.StdEncoding.EncodeToString(bpriv))

		},
	}
	return root
}

