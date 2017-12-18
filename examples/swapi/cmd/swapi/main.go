package main

import (
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
	"github.com/caicloud/nirvana/cli"
	"github.com/spf13/cobra"
	"os"
	"path"
	"github.com/caicloud/nirvana/examples/swapi/pkg/api"
	"net/http"
	"fmt"
)

const ErrCode = 1

func main() {
	var port uint
	var dp string
	c := cli.NewCommand(&cobra.Command{
		RunE: func(cmd *cobra.Command, _ []string) error {
			ml, err := loader.New(toAbs(dp))
			if err != nil {
				return err
			}
			s := api.CreateWebServer(ml)
			hostAndPort := fmt.Sprintf(":%d", port)
			fmt.Printf("start listening on %v\n", hostAndPort)
			http.ListenAndServe(hostAndPort, s)
			return nil
		},
	})

	c.AddFlag(
		cli.UintFlag{
			Name:        "port",
			Shorthand:   "p",
			Usage:       "port that the server listens to",
			Destination: &port,
			DefValue:    8000,
		},
		cli.StringFlag{
			Name:        "data-path",
			Shorthand:   "d",
			Usage:       "supply the data path",
			Destination: &dp,
		},
	)

	if err := c.Execute(); err != nil {
		os.Exit(ErrCode)
	}
}

func toAbs(dataPath string) string {
	if path.IsAbs(dataPath) {
		return dataPath
	} else if p, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		return path.Join(p, dataPath)
	}
}
