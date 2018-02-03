// Package post hat Funktionen für das Posten von Daten
package post

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var influxurl string

// Posterinflux schreibt nach Influx mit einem io-reader
func Posterinflux(anfrage io.Reader) error {
	if len(influxurl) == 0 {
		panic("influx nicht gesetzt")
	}
	resp, err := http.Post(influxurl, "application/text", anfrage)

	if err != nil {
		return errors.Wrapf(err, "postinflux")
	}

	io.Copy(ioutil.Discard, resp.Body) // <= NOTE antwort wird gelesen und ignoriert

	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		logrus.WithFields(logrus.Fields{
			"code":   resp.StatusCode,
			"status": resp.Status,
			"len":    resp.ContentLength,
		}).Error("influx_liefert_code")

	} else if resp.StatusCode >= 400 { // ist ein Fehler
		return fmt.Errorf("influx statuscode %d %s", resp.StatusCode, resp.Status)

	}

	return nil
}

func init() {
	var ok bool
	var influxhost string

	influxhost, ok = os.LookupEnv("PU_INFLUX_HOST")
	if ok {
		influxurl = fmt.Sprintf(`%s://%s:%s@%s:%s/write?db=%s&precision=s`,
			os.Getenv("PU_INFLUX_PROTOCOL"),
			os.Getenv("PU_INFLUX_USER"),
			os.Getenv("PU_INFLUX_PASSWORD"),
			influxhost,
			os.Getenv("PU_INFLUX_PORT"),
			os.Getenv("PU_INFLUX_DATABASE"))
	}

}
