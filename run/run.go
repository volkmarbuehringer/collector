// Package run frägt zyklisch server ab, die url werden aus der DB ausgelesen
package run

import (
	"context"
	"time"

	"collector/workparam"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type StatOut struct {
	Startrunde         time.Time `json:"zeit  systemstart"`
	Run                int       `json:"laufende runde"`
	Rundenstart        time.Time `json:"zeit rundenstart"`
	Rundenstartletz    time.Time `json:"zeit letzten rundenstart"`
	Rundensekunden     int       `json:"sekunden seit rundenstart"`
	Rundensekundenletz int       `json:"Dauer letzte Runde"`
	Rundendurchschnitt float64   `json:"Rundendurchschnitt"`
	Indexer            int       `json:"gestartete Abfragen"`
	Beendet            int       `json:"beendete Abfragen"`
	RunLen             int       `json:"Server in letzter Runde"`
	ReturnLen          int       `json:"Server in dieser Runde"`
	Performance        uint64    `json:"gelesene xml-records"`
	Durchschnittwert   float64   `json:"durchschnittl. laufzeit in ms"`
	Langsamzahl        int       `json:"laufzeit länger wie 30sec in runde"`
	Gleiche            int       `json:"server mit maximaler rundenzahl"`
	DurchschnittRunde  uint      `json:"Durchschnittzeit diese Runde"`
	PerformanceRunde   uint64    `json:"xml-Records diese Runde"`
	BeendetRunde       int       `json:"abgefragte Server diese Runde"`
	PerformanceSec     int       `json:"Performance/sec diese Runde"`
	Performanceletz    uint64    `json:"xml-Records vor Runde"`
	Beendetletz        int       `json:"abgefragte server vor Runde"`
	Fehlerletz         int       `json:"Fehler vor Rund"`
	Fehleranz          int       `json:"Fehler gesamt"`
	gesamtdurch        uint
	Laufende           int `json:"parallel laufende server"`
	Wartenzahl         int `json:"warten auf start"`
	rundenzeit         time.Duration
}

func (s *StatOut) reset() {
	s.Performanceletz = s.Performance
	s.Beendetletz = s.Beendet

	s.Langsamzahl, s.gesamtdurch = 0, 0
	s.Rundenstartletz = s.Rundenstart
	s.Rundenstart = time.Now()
	s.Rundensekundenletz = s.Rundensekunden
	s.Fehlerletz = s.Fehleranz
	t := s.Run
	if t > 5 {
		t = 5
	}
	if t > 0 {
		s.Rundendurchschnitt -= float64(s.Rundendurchschnitt) / float64(t)
		s.Rundendurchschnitt += float64(s.Rundensekundenletz) / float64(t)
	}

}

func (s *StatOut) compute() {

	if t := uint((s.Beendet - s.Beendetletz) - s.Langsamzahl); t > 0 {
		s.DurchschnittRunde = s.gesamtdurch / t
	}
	s.rundenzeit = time.Since(s.Rundenstart)
	s.Rundensekunden = int(s.rundenzeit/time.Second) + 1

	s.PerformanceRunde = s.Performance - s.Performanceletz
	s.BeendetRunde = s.Beendet - s.Beendetletz
	s.PerformanceSec = (int(s.Performance-s.Performanceletz) / s.Rundensekunden)

}

// StatInfo liefert die aktuellen Statistiken der Ausführung
type StatInfo struct {
	StatOut

	log *logrus.Logger
	ctx context.Context

	workparam.Loader                                  //interface zum laden der daten
	idmap            map[int]workparam.ParamStore     //zum checken ob server schon da sind
	tmpmap           map[workparam.StoreWork]struct{} //für gelöschte zu finden
	searchmap        map[workparam.Param]struct{}
	rundaten         []workparam.Param // aktuelle ausführung
	returndaten      []workparam.Param // ausführung nächste runde
	merkerdaten      []workparam.Param //gemerkte, die im moment nicht ausgeführt werden

	channelwait  time.Duration
	ziel         int
	started      int
	c            chan workparam.Param //zum starten
	returnsofort chan workparam.Param // für die ergebnisse

	maxserver int
	g         *errgroup.Group
	parallel  int
}

func (r *StatInfo) startworker() bool {

	if r.started < r.parallel { //nur solgang noch nicht maximale parallelität erreicht ist
		//starten des workers innerhalb einer errorgroup
		r.g.Go(r.workerrun)
		r.started++

		return true
	}
	return false

}

// Runner liefert einen Funktion, die die eigentliche Abfrage-Arbeit durchführt
// für die Initialisierung muss die Worker-Funktion geholt werden, die jedesmal für jeden server
// aufgerufen werden muss

//hauptfunktion: schleife n-mal durchführen, hier werden die runden über die
//Daten durchgeführt, es müssen ergebnisse in der runde erzielt werden
//bei fehler wird abgebrochen
