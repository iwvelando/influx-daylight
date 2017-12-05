package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/nathan-osman/go-sunrise"
	"github.com/urfave/cli"
)

func nextSunriseSunset(latitude, longitude float64, t time.Time) (time.Time, time.Time) {
	sunriseTime, sunsetTime := sunrise.SunriseSunset(
		latitude,
		longitude,
		t.Year(),
		t.Month(),
		t.Day(),
	)
	if t.After(sunriseTime) {
		t = t.Add(24 * time.Hour)
		return sunrise.SunriseSunset(
			latitude,
			longitude,
			t.Year(),
			t.Month(),
			t.Day(),
		)
	} else {
		return sunriseTime, sunsetTime
	}
}

func createBatchPoints(database string, sunriseTime, sunsetTime time.Time) (client.BatchPoints, error) {
	sunrisePt, err := client.NewPoint(
		"daylight",
		nil,
		map[string]interface{}{"value": 1},
		sunriseTime,
	)
	if err != nil {
		return nil, err
	}
	sunsetPt, err := client.NewPoint(
		"daylight",
		nil,
		map[string]interface{}{"value": 0},
		sunsetTime,
	)
	if err != nil {
		return nil, err
	}
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	if err != nil {
		return nil, err
	}
	bp.AddPoints([]*client.Point{sunrisePt, sunsetPt})
	return bp, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "influxdl"
	app.Usage = "Insert sunrise and sunset points into InfluxDB"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "influxdb-addr",
			Value:  "http://localhost:8086",
			EnvVar: "INFLUXDB_ADDR",
			Usage:  "address of InfluxDB server",
		},
		cli.StringFlag{
			Name:   "influxdb-username",
			EnvVar: "INFLUXDB_USERNAME",
			Usage:  "username for InfluxDB server",
		},
		cli.StringFlag{
			Name:   "influxdb-password",
			EnvVar: "INFLUXDB_PASSWORD",
			Usage:  "password for InfluxDB server",
		},
		cli.StringFlag{
			Name:   "influxdb-database",
			EnvVar: "INFLUXDB_DATABASE",
			Usage:  "database for InfluxDB server",
		},
		cli.Float64Flag{
			Name:   "latitude",
			EnvVar: "LATITUDE",
			Usage:  "location for calculating times",
		},
		cli.Float64Flag{
			Name:   "longitude",
			EnvVar: "LONGITUDE",
			Usage:  "location for calculating times",
		},
	}
	app.Action = func(c *cli.Context) error {

		log.Print("started")
		defer log.Print("stopped")

		// Create an InfluxDB client
		cl, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:     c.String("influxdb-addr"),
			Username: c.String("influxdb-username"),
			Password: c.String("influxdb-password"),
		})
		if err != nil {
			return err
		}
		defer cl.Close()

		// Make a channel for receiving a SIGINT or SIGTERM
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Loop until the next time occurs or a signal is received
		for {
			sunriseTime, sunsetTime := nextSunriseSunset(
				c.Float64("latitude"),
				c.Float64("longitude"),
				time.Now(),
			)
			log.Printf("waiting for %s", sunriseTime.String())
			select {
			case <-time.After(time.Until(sunriseTime)):
				bp, err := createBatchPoints(
					c.String("influxdb-database"),
					sunriseTime,
					sunsetTime,
				)
				if err != nil {
					log.Print(err.Error())
					break
				}
				log.Print("inserting points...")
				if err := cl.Write(bp); err != nil {
					log.Print(err.Error())
					break
				}
				continue
			case <-sigChan:
				return nil
			}
			log.Print("30 second timeout")
			select {
			case <-time.After(30 * time.Second):
			case <-sigChan:
				return nil
			}
		}
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
