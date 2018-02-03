



#einstellungen für programm
export PU_DEBUG=0 ##debugging ein
export PU_RUNDEBUG=1 ##zusätzliche meldungen bei run
export PU_LOGLEVEL=warn

export PU_PARSEDEBUG=0  ##debugging ein für parsen
export PU_STAT_INFO=0  ##ausgeben von statinfo per ticker, 0 abgeschaltet
export PU_MIN_REPEAT_TIME=6 ##wartezeit in sekunden nach abschluss der anfrage vor nächster anfrage
export PU_MIN_RUNDEN_TIME=7

export PU_INFLUX_BUFFER=100000  ##grösse des influx-buffer in bytes
export PU_INFLUX_WORKER=4  ## anzahl der influx-worker
export PU_SLOW_TIME=30,120,300,500 ##zeit in sekunden ab der nur jede n.te runde gemacht wird
export PU_SLOW_WORK=7,17,27,57 ## wie oft werden langsame ausgeführt 1,2,3 = jedes mal, jedes 2.mal usw

export PU_CHANNEL_WAIT=1100  ##maximale wartezeit auf channel in sekunden
export PU_DRY_RUN=10000 ##soviele runden ohne ergebnisse
export PU_START_DELAY=20 ## wartezeit in ms bis nächster worker gestartet wird
export PU_MAXMINUTEN=5 ##maximale zeit für xml-erzeugung
export PU_MAXSERVER=10000  ##maximale anzahl server

export PU_NEULADEN=1  ##wie oft werden die maps geladen (in runden 10= jede 10. runde)


export PU_MAXLANGSAME=50 #maximale anzahl langsame
export PU_MAX_RUNDEN_TIME=450
export PU_INFLUX_WAIT=10



#telegraf
export PU_INFLUX_HOST=localhost
export PU_INFLUX_PORT=8186
export PU_INFLUX_PROTOCOL="http"
export PU_INFLUX_USER=""
export PU_INFLUX_PASSWORD=""
export PU_INFLUX_DATABASE=""

export PORT=9000


##parameter: 1. Anzahl Worker
## 2. Anzahl runden bis ende
## 3. _id für test (0=alle)

go run  test/test.go $1 0 $2
