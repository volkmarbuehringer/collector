package workerbase

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func (f *ParamSub) GetLogger() *logrus.Entry {
	f.logger = logrus.WithFields(logrus.Fields{
		"scadanr":            f.ID,
		"url":                f.URL,
		"runs":               f.Runs,
		"runneu":             f.Runneu,
		"performance":        f.performance,
		"letzte_performance": f.letztPerformance,
		"letzte_laufzeit":    f.Duration,
		"durchschnitt":       f.Durchschnitt,
		"fehleranz":          f.Fehleranz,
	})
	return f.logger
}

type ParamRun struct {
	Starttime        time.Time
	letztPerformance uint64
	performance      uint64
	Durchschnitt     float64
	Duration         time.Duration
	Runs             int
	Runneu           int
	Fehleranz        int
	Letztfehler      bool
	Runflag          bool
}

func (work *ParamRun) Do() error {
	if work.Runflag {
		return fmt.Errorf("interner fehler runmap start")
	}

	work.Runs++ //run erhöhen
	work.Starttime = time.Now()
	work.Runflag = true
	work.Letztfehler = false
	return nil
}
func (work *ParamSub) RunCheck() bool {

	work.Runneu--
	return work.Runneu <= 0
}

func (work *ParamRun) SetPerformance(lens uint64) {
	work.letztPerformance = lens
	work.performance += lens

}

func (w *ParamRun) Ende() {
	w.Duration = time.Since(w.Starttime)
	w.Runflag = false
}

// Param sind die parameter, die der Worker-funktion übergeben werden
type ParamSub struct {
	ParamRun
	ID  int
	URL string

	Geloescht bool

	logger *logrus.Entry
}

func (p *ParamSub) GetID() int {
	return p.ID
}
func New(id int, url string, geloescht bool, dur time.Duration) ParamSub {

	t := ParamSub{ID: id, URL: url, Geloescht: geloescht,
		ParamRun: ParamRun{Duration: dur}}

	t.logger = t.GetLogger()
	return t
}
