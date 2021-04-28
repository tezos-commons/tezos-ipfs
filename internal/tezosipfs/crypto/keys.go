package crypto

import (
	"encoding/base64"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/db"
	"os"
)

func GetPrivateKey(db *db.BoltDb, l *logrus.Entry) []byte {
	log := l.WithField("source","config")
	if val,ok := os.LookupEnv("P2P_SECRETKEY"); ok {
		log.Info("Using private key from env")
		valb,_ := base64.StdEncoding.DecodeString(val)
		return valb
	}

	val,ok := db.Get([]byte("Config"),[]byte("libp2p_private_key"))
	if ok >= 1 {
		log.Info("Using private key from BoltDb")
		return val
	}

	// make new key and save in db
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	
	if err != nil {
		panic(err)
	}
	b,err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}
	db.Write([]byte("Config"),[]byte("libp2p_private_key"),b)
	log.Info("Generated new private key")
	return b
}
