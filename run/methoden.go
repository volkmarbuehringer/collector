package run

import (
	"fmt"
	"time"

	"collector/workparam"

	"github.com/sirupsen/logrus"
)

func (s *StatInfo) checkLangsame() int {
	var merkerlen = len(s.merkerdaten)

	if s.rundenzeit < maxrundentime {
		helper := make([]workparam.Param, 0)

		for _, v := range s.merkerdaten {
			if _, ok := s.searchmap[v]; ok {
				if v.RunCheck() && merkerlen-len(helper) <= maxlangsame {
					s.returndaten = append(s.returndaten, v)
				} else {
					helper = append(helper, v)
				}
			}
		}
		s.merkerdaten = helper
	}
	return merkerlen
}

func (s *StatInfo) rundenwechsel() error {
	//aktueller systemzustand
	//überprüfen ob überhaupt irgendein ergebnis in der runde erzielt wurde
	//warten auf ergebnisse

	if nachladen > 0 && s.Run%nachladen == 0 {
		if err := s.addServer(); err != nil {
			return err
		}

	}

	s.compute()

	var merkerlen = s.checkLangsame()
	rundenfehler := s.Fehleranz - s.Fehlerletz
	s.log.WithFields(logrus.Fields{
		"index":                s.Indexer,
		"runde":                s.Run,
		"gleiche_runde":        s.Gleiche,
		"daten_len":            len(s.idmap),
		"return_channel":       len(s.returnsofort),
		"laufende":             s.Laufende,
		"rundelen_alt":         len(s.rundaten),
		"rundelen_neu":         len(s.returndaten),
		"runden_dauer":         s.rundenzeit,
		"laufzeit_gesamt":      time.Since(s.Startrunde),
		"zurückgestellte_len":  len(s.merkerdaten),
		"wieder_in_runde":      merkerlen - len(s.merkerdaten),
		"ausgeführte_langsame": s.Langsamzahl,
		"fehler_in_runde":      rundenfehler,
	}).Info("info_zu_rundenende")

	//warte schleife für rundenende
	if err := s.statfun(); err != nil {
		return err
	}
	s.compute()

	if s.Performance == s.Performanceletz && s.Run%dryrun >= dryrun-1 {
		return (fmt.Errorf("kein Fortschritt"))
	}

	s.reset()

	s.log.WithFields(logrus.Fields{
		"beendet_gesamt":      s.Beendet,
		"performance_gesamnt": s.Performance,
		"laufende":            s.Laufende,
		"durchschnitt":        int(s.Durchschnittwert),
		"durchschnitt_runde":  s.DurchschnittRunde,
		"beendete":            s.BeendetRunde,
		"performance":         s.PerformanceRunde,
		"performance_pro_sek": s.PerformanceSec,
		"fehlerquote":         rundenfehler * 100 / s.BeendetRunde,
		"rundelen_neu":        len(s.returndaten),
	}).Info("vergleich_mit_letzter_runde")

	return nil
}

func (w *StatInfo) addServer() error {

	if w.Run > 0 {
		//maps neu laden

		//bei der nächsten ausführung wird map neu ermittelt
		w.tmpmap = make(map[workparam.StoreWork]struct{}, w.maxserver)
		for _, v := range w.idmap {
			w.tmpmap[v] = struct{}{}
		}
	}

	if err := w.AddServer(w); err != nil {
		return err
	}

	if len(w.returndaten) == 0 {
		return fmt.Errorf("keine Server für Abfrage selektiert")
	} else if w.Run == 0 {

		logrus.WithFields(logrus.Fields{
			"server_anzahl": len(w.returndaten),
		}).Info("start-setup")
		w.rundaten, w.returndaten = w.returndaten, nil
	} else {
		for key := range w.tmpmap {
			if err := key.Deleter(); err != nil {
				return err
			}
			if v, ok := w.idmap[key.GetID()]; ok {
				delete(w.idmap, key.GetID())
				delete(w.searchmap, v)
			}

		}
	}
	return nil
}
