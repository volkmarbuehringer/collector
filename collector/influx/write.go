// Package influx macht den Setup für die Influx-Schreibe Funktion und liefert eine Funktion zurück mit der in
// den Channel geschrieben wird
package influx

import (
	"bytes"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//Write  um den Influx-Channel zu füllen
//funktion für das reinschreiben von daten in den influx channel
//von slice mit xml-input daten
func (r *Influxer) Write(fun func(buffer *bytes.Buffer) error) error {

	timer := time.NewTimer(time.Duration(r.influxwait) * time.Millisecond)
	var start time.Time

	merk := r.wartenzahl
	var i int
	for { //zwei versuche

		select {
		case retwork, ok := <-r.returnsofort:
			if !ok {
				return fmt.Errorf("kanal zu")
			}
			r.gesamtlength += retwork
			r.schreibzahl++

		case r.influxchan <- fun: //ok
			timer.Stop()
			if i > 0 && r.started == r.workeranz {

				if merk+r.started-1 == r.wartenzahl {
					t := time.Since(start)
					r.GetLogger().WithFields(logrus.Fields{
						//		"Len":             len(xmldat),
						"Verzögerung_akt": t,
					}).Warn("langes Warten auf Influx")
					r.warnzahl++
				} else {
					r.wartenzahl++
				}

			}
			r.channelzahl++
			return nil //alles ok

		case start = <-timer.C:

			if i > 0 { //zweiter versuch mit langem warten schiefgegangen
				return errors.Wrapf(errors.New("influxlangsam"), "%v len %d", start, r.gesamtlength)
			}

			i++

			//starte neue worker
			if r.started < r.workeranz {
				r.g.Go(r.worker)
				r.started++

			}

			timer.Reset(30 * time.Second) //für zweiten versuch lange warten

		}

	}

}
