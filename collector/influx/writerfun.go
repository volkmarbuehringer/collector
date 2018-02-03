// Package influx macht den Setup für die Influx-Schreibe Funktion und liefert eine Funktion zurück mit der in
// den Channel geschrieben wird
package influx

import (
	"bytes"
	"context"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

//Influxer für die Ausführung der Influx writer
type Influxer struct {
	buffersize   int
	workeranz    int
	started      int
	wartenzahl   int
	gesamtlength int
	schreibzahl  int
	influxwait   int
	warnzahl     int
	channelzahl  int
	influxchan   chan func(*bytes.Buffer) error
	g            *errgroup.Group
	ctx          context.Context
	returnsofort chan int
}

func (r *Influxer) GetLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"workeranz":     r.workeranz,
		"started":       r.started,
		"wartenzahl":    r.wartenzahl,
		"channelwrites": r.channelzahl,
		"warnzahl":      r.warnzahl,
		"gesamtlength":  r.gesamtlength,
		"schreibzahl":   r.schreibzahl,
	})

}

//Init das gesamte Setup und das Einlesen der Daten in Maps wird hier durchgeführt
//input: der context
//output:  die struct
func Init(ctx context.Context) (*Influxer, error) {
	var r Influxer

	r.buffersize, _ = strconv.Atoi(os.Getenv("PU_INFLUX_BUFFER"))
	r.workeranz, _ = strconv.Atoi(os.Getenv("PU_INFLUX_WORKER"))
	r.influxwait, _ = strconv.Atoi(os.Getenv("PU_INFLUX_WAIT"))
	r.influxchan = make(chan func(*bytes.Buffer) error)
	r.returnsofort = make(chan int, 2000)

	r.g, r.ctx = errgroup.WithContext(ctx) //context verändern mit dieser errorgroup
	//starte ein worker
	r.g.Go(r.worker)
	r.started++

	return &r, nil
}

//Ende um Influx-Channel zu beenden funktion zum beenden des channels
func (r *Influxer) Ende() error {
	close(r.influxchan)

	if err := errors.Wrapf(r.g.Wait(), "influxgroup"); err != nil { //fehler kam von influx
		r.GetLogger().WithFields(logrus.Fields{
			"fehler": err,
		}).Error("Fehler von Influx")
		return err
	}

	close(r.returnsofort)
	r.GetLogger().WithFields(logrus.Fields{}).Warn("ende_influx")

	return nil

}
