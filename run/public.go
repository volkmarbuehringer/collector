package run

import (
	"fmt"

	"collector/workparam"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (w *StatInfo) Check(p workparam.ParamStore) (workparam.ParamStore, bool) {

	id := p.GetID()
	if serverid > 0 {
		if id != serverid {
			return nil, false
		}
	}

	f, ok := w.idmap[id]
	if ok {
		delete(w.tmpmap, f)
	} else {
		if len(w.idmap) >= w.maxserver {
			return f, false
		}
		w.returndaten = append(w.returndaten, p)
		w.idmap[p.GetID()] = p
		w.searchmap[p] = struct{}{}
	}
	return f, ok
}

func (r *StatInfo) SetParallel(parallel int) { //dies ist die endefunktion,system wird kontrolliert beendet
	if parallel > r.parallel {
		r.parallel = parallel
	}

}

func (run *StatInfo) Statistik() interface{} {
	t := run.StatOut
	t.compute()
	return t
}

// eine runde auf den selektierten servern die work funktion ausführen

func (run *StatInfo) Runner(anz int) error {

	if err := run.addServer(); err != nil {
		return err
	}
	run.reset()
	run.ReturnLen = len(run.rundaten)

	//hier ist die eigentliche ausführungsschleife
	for runner := 0; runner < anz; runner++ { //endlosschleife immer wieder über die datenbasis

		if len(run.rundaten) == 0 {
			return fmt.Errorf("keine daten %d", run.Run)
		}

		if err := errors.Wrapf(run.starter(run.rundaten), "startfehler %d", run.Indexer); err != nil {
			return err
		}

		if err := run.rundenwechsel(); err != nil {
			return err
		}
		run.ReturnLen, run.RunLen = len(run.returndaten), len(run.rundaten)

		run.rundaten, run.returndaten = run.returndaten, nil

	}
	return nil
}

func (r *StatInfo) Ende() error { //dies ist die endefunktion,system wird kontrolliert beendet
	close(r.c) //channel schliessen, warten auf ende
	//dies ist ein fehler aus dem worker (background)

	err := errors.Wrapf(r.g.Wait(), "group") //hier kommt der fehler raus am schluss , wenn alles zuende
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"laufende": r.started,
			"stack":    err,
		}).Error("fehler von worker")

	}
	close(r.returnsofort)
	return err

}
