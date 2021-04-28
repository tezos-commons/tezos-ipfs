package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/tezoscommons/tezos-ipfs/internal/tezosipfs/app"
	"go.uber.org/dig"
	"os"
	"os/signal"
)

func GetRootCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "tipfs",
	}

	root.AddCommand(GetConfigCommand(c),GetRunCommand(c))
	root.AddCommand(GetToolsCommand(c))
	return root
}

func GetRunCommand(c *dig.Container) *cobra.Command {
	var root = &cobra.Command{
		Use:   "run",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.Invoke(func(g *app.Gateway) {
				if g != nil {
					go g.Run()
				}
			})

			if err != nil {
				fmt.Println(err)
			}

			var signal_channel chan os.Signal
			signal_channel = make(chan os.Signal, 1)
			signal.Notify(signal_channel, os.Interrupt)
			<-signal_channel
		},
	}

	root.AddCommand(GetConfigCommand(c))
	return root
}
