package run

import (
	"fmt"
	"time"

	"collector/workparam"

	"github.com/sirupsen/logrus"
)

func (s *StatInfo) statwork(retwork workparam.Param) {
	s.Laufende--
	duration, p, runzahl, fehlerflag, delay := retwork.EndTimer(s.Run)
	s.Performance += p
	s.Beendet++
	if fehlerflag {
		s.Fehleranz++
	}
	t := float64(s.Beendet)
	if t > 3000 {
		t = 3000
	}

	s.Durchschnittwert -= s.Durchschnittwert / t
	s.Durchschnittwert += float64(duration) / t

	if runzahl > s.Run {
		s.Run = runzahl
		s.Gleiche = 0
	}

	if runzahl == s.Run {
		s.Gleiche++
	}
	if time.Duration(duration)*time.Millisecond > 30*time.Second {
		s.Langsamzahl++
	} else {
		s.gesamtdurch += duration
	}
	if delay == 0 {

		s.returndaten = append(s.returndaten, retwork)

	} else {

		s.merkerdaten = append(s.merkerdaten, retwork)
		if debugger {
			retwork.GetLogger().WithFields(logrus.Fields{"gemerkte": len(s.merkerdaten)}).Debug("überspringe")

		}

	}

}

func (r *StatInfo) workerrun() error {
	//dies ist der code der parallel läuft
	var err error
looper:
	for { //endlos

		select {
		case <-r.ctx.Done(): //hier kommt der cancel
			break looper

		case daten, ok := <-r.c:
			if !ok {
				break looper
			}
			err = daten.Do()
			if err != nil {
				break looper
			}

			daten.Ende()
			select {
			case <-r.ctx.Done(): //hier kommt der cancel
				break looper
			case r.returnsofort <- daten:

			}

		}
	}
	return err
}

func (s *StatInfo) statfun() error {
	timer := time.NewTimer(minrundentime - time.Since(s.Rundenstart))
	for {
		select {
		case <-s.ctx.Done(): //hier kommt der cancel
			return s.ctx.Err()
		case retwork, ok := <-s.returnsofort:
			if !ok {
				return fmt.Errorf("kanal zu")
			}

			s.statwork(retwork)

		case <-timer.C:
			return nil

		}

	}

}

func (r *StatInfo) starter(workdaten []workparam.Param) error {

	//dieser code läuft nicht parallel im hintergrund, sondern im vordergrund
	//immer nur einmal, er kommuniziert über den channel mit den workern
	// er überwacht die arbeit indem er prüft wie lange es dauert bis der
	//nächste in die verarbeitung kommt, dieser ist dann nicht fertig,
	//sondern lediglich gestartet!!!

	var timer = time.NewTimer(1 * time.Millisecond)

	for idx, work := range workdaten {
		if _, ok := r.searchmap[work]; ok {
			var i int
			var start time.Time

		looper:
			for { //endlos

				select {
				case <-r.ctx.Done(): //hier kommt der cancel
					return r.ctx.Err() // errors.Wrapf(, "runner %d %v", work.ID, start)

				case retwork := <-r.returnsofort:
					r.statwork(retwork)

				case r.c <- work: //hier wird der job gestartet

					break looper //ausgang job gestartet

				case start = <-timer.C: //hier wird kontrolliert, wie lange es bis zum start dauert

					if !r.startworker() { //noch eine chance mit langem warten geben
						i++
						//noch einen worker starten, wenn möglich
						if i > 1 { //zweite chance vorbei, abbruch

							return fmt.Errorf("langsam gestört pos %d proc %d %v", idx, r.started, r.channelwait)

						}
						r.Wartenzahl++
						timer.Reset(r.channelwait)

					}

				}

			}
			r.Laufende++

			r.Indexer++

			if !start.IsZero() {
				//warnung wenn es zu lange dauert
				t := time.Since(start)
				if r.started > 10 && t > 10*time.Second {

					logrus.WithFields(logrus.Fields{
						"laufende":       r.started,
						"gewartet":       t,
						"maximal_warten": r.channelwait,
						"wartenzahl":     r.Wartenzahl,
						"gestartet":      idx,
					}).Warn("langes Warten auf Queue")

				}
			}

			timer.Reset(time.Duration(r.ziel) * time.Millisecond)
		}
	}
	return nil
}
