package runtest

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"

	"strconv"
	"time"

	"collector/collector/influx"
	"collector/workerbase"
	"collector/workparam"

	"github.com/sirupsen/logrus"
)

type TestBase struct {
	workerbase.ParamSub
	start time.Time
	dur   int
}

type Test struct {
	TestBase
}

type Test1 struct {
	TestBase
}

type Test2 struct {
	TestBase
}

var ctx context.Context

var influxer *influx.Influxer

type Instance struct {
}

func Init(ctxin context.Context) (*Instance, error) {
	var instance Instance
	ctx = ctxin
	if t, ok := os.LookupEnv("PU_DEBUG"); ok {
		debug, _ := strconv.ParseBool(t)
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

	}
	var err error
	influxer, err = influx.Init(ctxin)
	if err != nil {
		//setup bei influx gescheitert
		return nil, err
	}
	return &instance, nil
}

func Exit() {
	fmt.Println("hier ende")
	influxer.Ende()
}

func (r *Instance) AddServer(w workparam.Storer) error {
	for idx := 1; idx < 5000; idx++ {
		var t = New()

		if f, ok := w.Check(t); ok {
			ff := f.(*Test)
			ff.Change(t.ParamSub)

		}
		var t1 = New1()

		if f, ok := w.Check(t1); ok {
			ff := f.(*Test1)
			ff.Change(t1.ParamSub)

		}
		var t2 = New2()

		if f, ok := w.Check(t2); ok {
			ff := f.(*Test2)
			ff.Change(t2.ParamSub)

		}
	}

	return nil
}

func (work *TestBase) Deleter() error {
	return nil
}
func (work *TestBase) Do() error {

	// funktion um die daten mittels post zu lesen, evtl. mehrere versuche machen
	// evtl muss vorher gewartet werden
	// auf cancel wird dabei geprüft
	//nach dem lesen per post wird noch das xml umgewandelt
	// erst nach der xml-umwandlung kann der body geschlossen werden, weil stream

	if err := work.ParamRun.Do(); err != nil {
		return err
	}

	time.Sleep(time.Duration(int(work.dur)) * time.Millisecond)

	//select {
	//	case <-ctx.Done(): //cancel
	//		return errors.Wrapf(ctx.Err(), "influxcancel")
	//default:
	work.SetPerformance(uint64(rand.Intn(5)))
	if rand.Float64() < 0.1 {
		work.Letztfehler = true
	} else {
		var fun = func(buffer *bytes.Buffer) error {
			for i := 0; i < 10; i++ {
				buffer.WriteString("cpu_load_short,host=server01,region=us-west value=0.64 1434055562000000000\n")
			}

			return nil
		}

		return influxer.Write(fun)
		//	}

	}

	return nil

}

func NewBase(offset int) TestBase {
	dur := 2000.0 / rand.NormFloat64()
	if dur < 0 {
		dur = dur * -1
	}
	if dur > 60000 {
		dur = 60000
	}
	return TestBase{
		ParamSub: workerbase.New(offset+rand.Intn(5000), "RandStringBytesMaskImprSrc(30)", false, 0),
		dur:      int(dur),
	}
}
func New() *Test {

	return &Test{TestBase: NewBase(0)}

}

func New1() *Test1 {

	return &Test1{TestBase: NewBase(100000)}

}

func New2() *Test2 {

	return &Test2{TestBase: NewBase(200000)}

}
