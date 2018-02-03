package service

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo"
)

type RouteInter interface {
	SetParallel(int)
	Statistik() interface{}
}

type route struct {
	RouteInter
	echo.Context
	cancel func()
}

func cancelHandler(c echo.Context) error {
	cc := c.(*route)
	cc.cancel() //programm beenden

	return c.String(http.StatusOK, "ok")
}

func fasterHandler(c echo.Context) error {
	cc := c.(*route)
	x := c.QueryParam("degree")
	length, _ := strconv.Atoi(x)

	if len(x) == 0 {
		length = 10
	}

	cc.SetParallel(length)
	return c.String(http.StatusOK, "ok")
}

func getAllHandler(c echo.Context) error {
	cc := c.(*route)
	return c.JSON(http.StatusOK, cc.Statistik())
}

// Routing setzt den Router auf
func Routing(cancel func(), sp RouteInter) *http.Server {

	/*
		var getAllHandler1 = func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusOK)

				runs := 1 //stat.StatistikRuns()
				var sum int
				var anz int
				var anzs int
				var maxer int
				var maxkey int
				var minkey = 99999
				for key, v := range runs {
					if v > maxer {
						maxer = v
					}
					if key > maxkey {
						maxkey = key
					}
					if key < minkey {
						minkey = key
					}

					sum += v
				}
				langsamzahl := minkey + ((maxkey - minkey) / 4)
				schnellzahl := maxkey - ((maxkey - minkey) / 4)
				for key, v := range runs {
					if key <= langsamzahl {
						anz += v
					}
					if key >= schnellzahl {
						anzs += v
					}

				}

				if sum > 0 {
					var erg = map[string]int{
						"maximale anzahl":      maxer,
						"maximale runs":        maxkey,
						"minimale runs":        minkey,
						"langsamgrenze":        langsamzahl,
						"schnellgrenze":        schnellzahl,
						"gesamt":               sum,
						"langsame":             anz,
						"schnelle":             anzs,
						"anteil langsame in %": int((float64(anz) / float64(sum)) * 100.0),
						"anteil schnelle in %": int((float64(anzs) / float64(sum)) * 100.0),
					}
					if err := json.NewEncoder(w).Encode(erg); err != nil {
						logrus.WithFields(logrus.Fields{
							"Fault": err,
						}).Error("json")

					}

				}

		}
	*/
	e := echo.New()
	e.HideBanner = true
	t, ok := os.LookupEnv("PORT")

	if !ok {

		t = "8080"
	}
	s := &http.Server{
		Addr:         ":" + t,
		ReadTimeout:  20 * time.Minute,
		WriteTimeout: 20 * time.Minute,
	}

	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			return h(&route{sp, c, cancel})
		}
	})

	e.GET("/faster", fasterHandler)
	e.POST("/cancel", cancelHandler)
	e.GET("/statistik", getAllHandler)

	go func() {
		e.Logger.Fatal(e.StartServer(s))
	}()

	return s
}
