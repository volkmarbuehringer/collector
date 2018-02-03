package workerbase

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func (w *ParamRun) EndTimer(run int) (uint, uint64, int, bool, int) {

	merker := float64(w.Duration / time.Millisecond)
	t := float64(w.Runs)
	if t > 10.0 {
		t = 10.0
	}
	w.Durchschnitt -= w.Durchschnitt / t
	w.Durchschnitt += merker / t
	if w.Letztfehler {
		w.Fehleranz++
	}

	var temp time.Duration

	if w.Runs > 6 {
		var fehlerquote = float64(w.Fehleranz) / float64(w.Runs)
		temp = time.Duration(w.Durchschnitt) * time.Millisecond
		if fehlerquote < 0.3 && w.Duration < time.Duration(w.Durchschnitt/3)*time.Millisecond {
			temp = time.Duration(w.Durchschnitt/3) * time.Millisecond
		}

	} else {
		temp = w.Duration

	}
	anz := runModulus(temp)

	var kor int
	if run > 6 {
		okruns := w.Runs - w.Fehleranz

		if float64(okruns)*1.4 > float64(run) {
			kor = -5
		}

		if float64(okruns)/float64(w.Runs) < 0.3 {
			kor = 9
		}
	}
	if anz+kor > 1 {
		w.Runneu = anz + kor
	} else {
		w.Runneu = 0
	}

	return uint(merker),
		w.letztPerformance, w.Runs, w.Letztfehler, w.Runneu
}

//checken entsprechend laufzeit ob ausgeführt wird in dieser runde
func runModulus(duration time.Duration) int {
	vgl := int(duration / time.Second)
	if len(slowwork) == 0 {
		return 1
	}
	for i := range slowwork {
		if vgl < slowtime[i] {
			if i == 0 {
				return 1
			}
			return rand.Intn(slowwork[i-1]) + 1

		}
	}

	return rand.Intn(len(slowwork)-1) + 1

}

var startminuten int32
var slowwork []int
var slowtime []int
var maxminuten int32

func init() {
	ts1, ok1 := os.LookupEnv("PU_SLOW_WORK")
	ts2, ok2 := os.LookupEnv("PU_SLOW_TIME")

	if ok1 && ok2 {
		tt1 := strings.Split(ts1, ",")
		tt2 := strings.Split(ts2, ",")
		if len(tt1) != len(tt2) {
			panic("falsche argumente slow")
		}
		for i := range tt1 {
			t1, _ := strconv.Atoi(tt1[i])
			t2, _ := strconv.Atoi(tt2[i])
			slowwork = append(slowwork, t1)
			slowtime = append(slowtime, t2)

		}

	}

	s, _ := strconv.Atoi(os.Getenv("PU_STARTMINUTEN"))
	startminuten = int32(s)
	if startminuten < 1 {
		startminuten = 600
	}

	m, _ := strconv.Atoi(os.Getenv("PU_MAXMINUTEN"))
	maxminuten = int32(m)
	if maxminuten < 1 || maxminuten > 100 {
		maxminuten = 60
	}
}
