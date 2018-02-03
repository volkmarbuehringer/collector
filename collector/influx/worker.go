// Package influx macht den Setup für die Influx-Schreibe Funktion und liefert eine Funktion zurück mit der in
// den Channel geschrieben wird
package influx

import (
	"bytes"
	"time"

	"collector/post"

	"github.com/sirupsen/logrus"
)

/* parallel ausgeführter worker, der nach influx schreibt */
func (r *Influxer) worker() error {

	var buffer bytes.Buffer
	buffer.Grow(r.buffersize * 3)
	var dur time.Duration
	if r.started == 0 {
		dur = 999999999 * time.Second
	} else {
		dur = 30 * time.Second
	}
	timer := time.NewTimer(dur)
looper:
	for { //warten auf channel
		select {
		case <-timer.C: //
			break looper
		case fun, ok := <-r.influxchan:
			if !ok {
				break looper
			}
			if err := fun(&buffer); err != nil {
				r.GetLogger().WithFields(logrus.Fields{
					"Fehler": err,
				}).Error("Fehler Influx Parser")
				return err

			} else if buffer.Len() > r.buffersize { //buffer voll jetzt schreiben
				var length = buffer.Len()
				if err := post.Posterinflux(&buffer); err != nil {
					r.GetLogger().WithFields(logrus.Fields{
						"Len":    length,
						"Fehler": err,
						//	"scadanr": tt.scadanr,
					}).Error("Fehler Influx Post")

					return err
				} else {
					r.returnsofort <- length
					timer.Reset(dur)

				}

			}
		}

	}

	//buffer flushen zum schluss
	if len := buffer.Len(); len > 0 {

		if err := post.Posterinflux(&buffer); err != nil {
			r.GetLogger().WithFields(logrus.Fields{
				"Len":    len,
				"Fehler": err,
			}).Error("Fehler Flush")
			return err
		}
		r.returnsofort <- len
	}
	r.started--

	return nil
}
