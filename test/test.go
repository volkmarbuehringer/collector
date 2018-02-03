// Package main, Ausführung opc-Server Abfragen
// System wird vorbereitet und der eigentliche Runner aufgerufen
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"collector/run"
	"collector/runtest"
	"collector/service"
	"collector/workparam"
)

func main() {

	if len(os.Args) < 4 {
		panic(fmt.Errorf("zuwenig argumente"))
	}
	parallel, err := strconv.Atoi(os.Args[1])

	if err != nil || parallel < 1 || parallel > 5000 {
		panic(fmt.Errorf("zuwenig argumente"))
	}

	anz, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(fmt.Errorf("zuwenig argumente"))
	}

	if anz == 0 {
		anz = 999999999999
	}

	if err := mainwork(parallel, anz); err != nil {
		logrus.WithFields(logrus.Fields{
			"stack": fmt.Sprintf("%+v", err),
		}).Fatal("abbruch")
	}

}

func mainwork(parallel int, anz int) error {

	//für den gesamtcontext wird ein cancel geholt
	ctx, cancel := Contexter()

	var fun workparam.Loader

	fun, err := runtest.Init(ctx)

	if err != nil {

		return err
	}
	/*
		//setup work funktion vorbereiten mit dem veränderten context der influx-errorgroup
		setup.Workfun, err = worker.GetWorker(influxer)

	*/

	//defer influxer.Ende() //influx flushen, wenn ein fehler aus influx kam, wird er erst hier geholt

	start := time.Now()

	// durchführung des laufs mit hilfe der übergegebenen funktionen starter und workfun
	collect := run.MakeWorkdaten(ctx, fun, parallel)
	h := service.Routing(cancel, collect) //service kann stat anzeigen und canceln

	defer h.Shutdown(ctx)

	if err := collect.Runner(anz); err != nil {
		logrus.WithFields(logrus.Fields{
			"dauer": time.Since(start),
			"stack": err,
		}).Error("fehler beim starten")
		cancel() //cancel drücken, für alle fälle
		collect.Ende()
		runtest.Exit()
		return err
	}
	//hier ist das ende erreicht und es muss noch alles heruntergefahren werden
	logrus.WithFields(logrus.Fields{
		"dauer": time.Since(start),
	}).Info("Lauf OK.Gesamtzeit")
	return nil

}

//Contexter installiert ein signal-handler und den cancel-Context
func Contexter() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		for {
			s := <-signal_chan
			switch s {
			// kill -SIGHUP XXXX
			case syscall.SIGHUP:
				fmt.Println("hungup")
				cancel()

			// kill -SIGINT XXXX or Ctrl+c
			case syscall.SIGINT:
				fmt.Println("Warikomi")
				cancel()

			case syscall.SIGQUIT:
				fmt.Println("stop and core dump")
				cancel()

			// kill -SIGTERM XXXX
			case syscall.SIGTERM:
				fmt.Println("force stop")
				cancel()

			default:
				fmt.Println("Unknown signal.")
			}
		}
	}()
	return ctx, cancel
}
