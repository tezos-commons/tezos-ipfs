package main

import (
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"github.com/tezoscommons/tezos-ipfs/cmd"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/app"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/cache"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/config"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/crypto"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/db"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/network"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/swarm"
	"go.uber.org/dig"
	"gopkg.in/sohlich/elogrus.v7"
	"io"
	"log"
	"os"
	"time"
)

func main(){
	c := dig.New()
	c.Provide(config.NewConfig)
	c.Provide(network.NewIPFS)
	c.Provide(network.NewLightclient)
	c.Provide(db.NewStormDB)
	c.Provide(crypto.GetPrivateKey)
	c.Provide(cache.NewS3Cache)
	c.Provide(app.NewGateway)
	c.Provide(network.GetNetwork)
	c.Provide(swarm.NewSwarm)
	c.Provide(GetLog)
	c.Provide(app.NewPinManager)
	c.Provide(app.NewAdminAPI)

	rootCmd := cmd.GetRootCommand(c)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}


}


func GetLog (c *config.Config) *logrus.Entry{

	l := logrus.New()
	if c.Log.Elasticsearch != "" {
		client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
		if err != nil {
			log.Fatal(err)
		}
		hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, "tezos-ipfs")
		if err != nil {
			log.Fatal(err)
		}
		l.AddHook(hook)
	}

	if c.Log.File != "" {
		logFile,e := os.OpenFile(c.Log.File,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if e != nil {
			fmt.Println(e)
		}
		mw := io.MultiWriter(os.Stdout, logFile)
		l.SetOutput(mw)
		l.SetLevel(logrus.TraceLevel)
	}
	if c.Log.Format == "text" {
		l.SetFormatter(&logrus.TextFormatter{})
	}
	if c.Log.Format == "json" {
		l.SetFormatter(&logrus.JSONFormatter{})
	}
	return l.WithField("starttime",time.Now().Unix())
}