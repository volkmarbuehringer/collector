// Package run frägt zyklisch server ab, die url werden aus der DB ausgelesen
package run

import (
	"context"
	"os"
	"strconv"
	"time"

	"collector/workparam"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var debugger bool

var dryrun int
var minrundentime time.Duration
var maxrundentime time.Duration
var maxlangsame int
var nachladen int
var serverid int

func MakeWorkdaten(ctxin context.Context, fun workparam.Loader, parallel int) *StatInfo {
	t, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}
	serverid = t

	t, _ = strconv.Atoi(os.Getenv("PU_CHANNEL_WAIT"))
	if t < 5 {
		t = 5
	}

	t1, _ := strconv.Atoi(os.Getenv("PU_START_DELAY"))
	if t1 < 10 {
		t1 = 10
	}

	maxserver, _ := strconv.Atoi(os.Getenv("PU_MAXSERVER"))
	if maxserver < 1 {
		maxserver = 10000
	}

	run := &StatInfo{
		channelwait:  time.Duration(t) * time.Second,
		ziel:         t1,
		parallel:     parallel,
		c:            make(chan workparam.Param),
		merkerdaten:  make([]workparam.Param, 0),
		returnsofort: make(chan workparam.Param),

		StatOut: StatOut{Rundenstart: time.Now(),
			Startrunde: time.Now()},
		maxserver:   maxserver,
		Loader:      fun,
		idmap:       make(map[int]workparam.ParamStore, maxserver),
		searchmap:   make(map[workparam.Param]struct{}, maxserver),
		returndaten: make([]workparam.Param, 0, maxserver),
	}
	run.g, run.ctx = errgroup.WithContext(ctxin)

	run.log = logrus.New()
	if debug, _ := strconv.ParseBool(os.Getenv("PU_RUNDEBUG")); debug {
		run.log.SetLevel(logrus.DebugLevel)
	} else {
		run.log.SetLevel(logrus.InfoLevel)
	}

	return run
}

func init() {

	maxlangsame, _ = strconv.Atoi(os.Getenv("PU_MAXLANGSAME"))
	if maxlangsame < 3 {
		maxlangsame = 10
	}
	nachladen, _ = strconv.Atoi(os.Getenv("PU_NEULADEN"))
	//	pdebug, _ = strconv.ParseBool(os.Getenv("PU_PARSEDEBUG"))

	debugger, _ = strconv.ParseBool(os.Getenv("PU_DEBUG"))

	t, _ := strconv.Atoi(os.Getenv("PU_MIN_RUNDEN_TIME"))

	if t < 1 {
		t = 1
	}
	minrundentime = time.Duration(t) * time.Second

	t, _ = strconv.Atoi(os.Getenv("PU_MAX_RUNDEN_TIME"))

	if t < 1 {
		t = 1
	}
	maxrundentime = time.Duration(t) * time.Second
	dryrun, _ = strconv.Atoi(os.Getenv("PU_DRY_RUN"))
	if dryrun < 1 {
		dryrun = 1
	}

}
