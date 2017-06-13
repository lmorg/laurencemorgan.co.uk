package main

import (
	"golang.org/x/net/websocket"
	"strings"
) // "net/http"

//type RTChan chan string

type RTUpdate struct {
	Write *string
	URI   string
}

var real_time_chans map[string]map[string]chan *string
var real_time_update chan RTUpdate

func rtWebSockets(ws *websocket.Conn) {
	var (
		session Session
		//access_log AccessLog
	)

	defer func() {
		debugLog("[realtime] [rtWebSockets] [" + session.r.URL.Path + "] [" + session.ID + "] return (that's bad)")
		session.ResponseSize += len(session.Page.Content)
		_, err := ws.Write([]byte(session.Page.Content))
		isErr(&session, err, true, "Writing to websocket in defer func()", "rtWebSockets")
		ws.Close()
		dbClose(session.db) // if db connection isn't already closed...
		accessLog(&session) // log http requests
	}()

	// create session
	session = NewSession(nil /* http.ResponseWriter */, ws.Request(), &live_layout)
	session.RealTimeRequest = true

	// split the URL, but also add some blank fields (it's a bit inefficient, but it makes reading the values easier later on)
	session.Path = strings.Split(session.r.URL.Path, "/")
	session.File = session.Path[len(session.Path)-1]
	for i := len(session.Path); i < 5; i++ {
		session.Path = append(session.Path, "")
	}

	// open a database connection for that session
	var failed bool
	failed, session.db = dbOpen()
	if failed {
		pageTooManyConnections(&session)
		return

	} else {
		autoLogin(&session)

		// TODO: authorisation gubbins required
		ws.Write([]byte("connected"))
		debugLog("[realtime] [rtWebSockets] [" + session.r.URL.Path + "] [" + Itoa(session.User.ID) + "] websocket created")
		if len(real_time_chans[session.r.URL.Path]) == 0 {
			real_time_chans[session.r.URL.Path] = make(map[string]chan *string)
		}
		real_time_chans[session.r.URL.Path][session.ID] = make(chan *string)

		for {
			s := <-real_time_chans[session.r.URL.Path][session.ID]
			debugLog("[realtime] [rtWebSockets] [" + session.r.URL.Path + "] [" + Itoa(session.User.ID) + "] real time chans received")

			_, err := ws.Write([]byte(*s))
			if err != nil {
				debugLog("[realtime] [rtWebSockets] [" + session.r.URL.Path + "] [" + Itoa(session.User.ID) + "] " + err.Error())
				return
			}
			debugLog("[realtime] [rtWebSockets] [" + session.r.URL.Path + "] [" + Itoa(session.User.ID) + "] websocket written")

			session.ResponseSize += len(*s)
		}

	}

}

func realTimeUpdateManager() {
	debugLog("Starting realTimeUpdateManager....")
	real_time_chans = make(map[string]map[string]chan *string)
	real_time_update = make(chan RTUpdate)

	for {
		push := <-real_time_update
		debugLog("[realtime] [realTimeUpdateManager] real time update received")

		for uid, _ := range real_time_chans[push.URI] {
			//go func() { real_time_chans[push.URI][uid] <- &push.Write }()
			real_time_chans[push.URI][uid] <- push.Write
			debugLog("[realtime] [realTimeUpdateManager] real time update pushed")
		}
	}
}

func RealTimeUpdate(session *Session, section_name string, data *string) {
	if ENABLE_REAL_TIME_DESKTOP || ENABLE_REAL_TIME_MOBILE {
		debugLog("[realtime] [RealTimeUpdate] sending update")
		real_time_update <- RTUpdate{
			Write: data,
			URI:   "/" + section_name + "/" + Itoa(session.Page.Section.ID),
		}
	}
	debugLog("[realtime] [RealTimeUpdate] return")
}
