package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nathan-osman/go-sunrise"
	"github.com/urfave/cli"
)

const (
	sunriseString = "sunrise"
	sunsetString  = "sunset"
)

func nextTime(latitude, longitude float64) (time.Time, string) {
	now := time.Now()
	sunriseTime, sunsetTime := sunrise.SunriseSunset(
		latitude,
		longitude,
		now.Year(),
		now.Month(),
		now.Day(),
	)
	if sunriseTime.After(now) {
		return sunriseTime, sunriseString
	}
	if sunsetTime.After(now) {
		return sunsetTime, sunsetString
	}
	now = now.Add(24 * time.Hour)
	sunriseTime, _ = sunrise.SunriseSunset(
		latitude,
		longitude,
		now.Year(),
		now.Month(),
		now.Day(),
	)
	return sunriseTime, sunriseString
}

func addAnnotation(addr, username, password, text string, t time.Time) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	u.Path = "/api/annotations"
	b, err := json.Marshal(map[string]interface{}{
		"time":     t.Unix(),
		"isRegion": false,
		"text":     text,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Influx Daylight")
	if username != "" && password != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "influxdl"
	app.Usage = "Insert sunrise and sunset annotations into InfluxDB"
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

		// Make a channel for receiving a SIGINT or SIGTERM
		sigChan := make(chan os.Signal)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Loop until the next time occurs or a signal is received
		for {
			t, text := nextTime(c.Float64("latitude"), c.Float64("longitude"))
			select {
			case <-time.After(time.Until(t)):
				if err := addAnnotation(
					c.String("influxdb-addr"),
					c.String("influxdb-username"),
					c.String("influxdb-password"),
					text,
					t,
				); err != nil {
					log.Print(err.Error())
				}
			case <-sigChan:
				break
			}
		}

		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err.Error())
	}
}
