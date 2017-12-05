package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "influxdl"
	app.Usage = "Insert sunrise and sunset annotations into InfluxDB"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "latitude",
			EnvVar: "LATITUDE",
			Usage:  "location for calculating times",
		},
		cli.StringFlag{
			Name:   "longitude",
			EnvVar: "LONGITUDE",
			Usage:  "location for calculating times",
		},
	}
	app.Action = func(c *cli.Context) {

		// Wait for SIGINT or SIGTERM
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
	}
	app.Run(os.Args)
}
