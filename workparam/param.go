package workparam

import (
	"github.com/sirupsen/logrus"
)

type StoreWork interface {
	Deleter() error
	GetID() int
}
type ParamStore interface {
	//Updater() error
	Param
	StoreWork
}

type ParamMap interface {
	Ende()
	EndTimer(int) (uint, uint64, int, bool, int)
	Get() interface{}
}

type ParamDo interface {
	ParamMap
	Do() error
}

type Param interface {
	ParamDo

	RunCheck() bool
	GetLogger() *logrus.Entry
}

type Loader interface {
	AddServer(Storer) error
}

type Storer interface {
	Check(ParamStore) (ParamStore, bool)
}
