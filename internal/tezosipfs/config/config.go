package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"sync"
)

type Config struct {
	Gateway Gateway
	Log Log
	PinManager PinManager
	DB DB
	Peers Peers
	PinManagerEnabled bool
	GatewayEnabled bool
	log *logrus.Entry
	Identity Identity
	lock *sync.Mutex
	updates []chan*Config
}

func (c *Config) GetUpdates() chan *Config {
	ch := make(chan *Config)
	c.updates = append(c.updates, ch)
	go func() {
		ch <- c
	}()
	return ch
}

type IPFS struct {
	API string `yaml:"API"`
}

type CORS struct {
	AllowedDomains []string `yaml:"AllowedDomains"`
}

type Uploads struct {
	Enabled bool `yaml:"Enabled"`
	MaxSize int  `yaml:"MaxSize"`
}

type PinManager struct {
	API string `yaml:"API"`
}

type Yaml2Go struct {
	Gateway    Gateway    `yaml:"Gateway"`
	Log        Log        `yaml:"Log"`
	PinManager PinManager `yaml:"PinManager"`
	DB         DB         `yaml:"DB"`
}

type S3 struct {
	Region string `yaml:"Region"`
	Bucket string `yaml:"Bucket"`
	Secret string `yaml:"Secret"`
	Key    string `yaml:"Key"`
	Endpoint string `yaml:"Endpoint"`
	DisableSSL bool `yaml:"DisableSSL"`
}

type Backend struct {
	IPFS    IPFS           `yaml:"IPFS"`
}

type Log struct {
	File          string `yaml:"File"`
	Format        string `yaml:"Format"`
	Level         string `yaml:"Level"`
	Elasticsearch string `yaml:"Elasticsearch"`
}

type Gateway struct {
	CORS       CORS       `yaml:"CORS"`
	Uploads    Uploads    `yaml:"Uploads"`
	Server     Server     `yaml:"Server"`
	Storage    Storage    `yaml:"Storage"`
	Backend    Backend    `yaml:"Backend"`
}

type Server struct {
	Port         int            `yaml:"Port"`
	AccessTokens []AccessTokens `yaml:"AccessTokens"`
	UploadToken  []AccessTokens  `yaml:"UploadToken"`
}

type AccessTokens struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

type Storage struct {
	S3     S3     `yaml:"s3"`
	Folder string `yaml:"Folder"`
}

type DB struct {
	Storm string `yaml:"Storm"`
}

type Identity struct {
	Name         string `yaml:"Name"`
	Organization string `yaml:"Organization"`
	Contact      string `yaml:"Contact"`
	Comment      string `yaml:"Comment"`
}


// Peers
type Peers struct {
	CacheFor     []string `yaml:"CacheFor"`
	PinFor       []string `yaml:"PinFor"`
	TrustedPeers []string `yaml:"TrustedPeers"`
}




func NewConfig() *Config{
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/tezos-ipfs/")
	viper.AddConfigPath("$HOME/.tezos-ipfs")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	c := Config{}
	c.lock = &sync.Mutex{}
	c.updates = []chan *Config{}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
		c.lock.Lock()
		defer c.lock.Unlock()
		err = viper.Unmarshal(&c)
		if err != nil {
			fmt.Println("unable to decode into struct, %v", err)
			os.Exit(1)
		}

	})
	err = viper.Unmarshal(&c)
	if err != nil {
		fmt.Println("unable to decode into struct, %v", err)
		os.Exit(1)
	}
	return &c
}