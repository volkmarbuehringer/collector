package workerbase

import (
	"github.com/sirupsen/logrus"
)

func (w *ParamSub) Get() interface{} {
	return *w
}
func (w *ParamSub) Updater() error {

	return nil
}

func (f *ParamSub) Change(s ParamSub) {

	f.Geloescht = s.Geloescht
	if f.Geloescht {
		logrus.WithFields(logrus.Fields{
			"scada_nr": f.ID,
		}).Warn("Server gelöscht")

	}
	if len(s.URL) > 0 && f.URL != s.URL {
		f.URL = s.URL
	}
}
