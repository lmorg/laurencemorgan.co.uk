package main

import (
	//"crypto/tls"
	"flag"
	"fmt"
	"github.com/kardianos/osext"
	"golang.org/x/net/websocket"
	"net/http"
	"runtime"
	"time"
)

func main() {
	flag.StringVar(&PWD_CONFIG_PATH, "conf", defaultConfDir(), "Directory for config files")
	flag.Parse()

	quit := make(chan int)

	infoLog("Starting", CMS_NAME+", version", version.Version)
	loadEnvironmentFromConf("level 10 fireball.json", PWD_CONFIG_PATH)
	loadEnvironmentFromDB(&live_layout)

	runtime.GOMAXPROCS(CORE_GO_MAX_PROCS)

	if ENABLE_REAL_TIME_DESKTOP || ENABLE_REAL_TIME_MOBILE {
		go realTimeUpdateManager()
		http.Handle("/websocket/", websocket.Handler(rtWebSockets))
	}
	http.HandleFunc("/", pageHandler)

	go listenerHTTP()
	go listenerTLS()
	if !SITE_ENABLE_HTTP && !SITE_ENABLE_TLS {
		errLog(CMS_NAME, "is not set to listen on any ports.")
		errLog("You need to enabled either SITE_ENABLE_HTTP or SITE_ENABLE_TLS or both.")
		quit <- 1
	}

	time.Sleep(1 * time.Second)
	//secureDaemon()
	go cacheForumsManager()
	go cacheFileManager()
	go cacheUsersManager()
	<-quit
}

func defaultConfDir() string {
	path, err := osext.ExecutableFolder()
	failOnErr(err, "defaultConfDir")
	return path + `/../conf/`
}

func listenerHTTP() {
	if SITE_ENABLE_HTTP {
		infoLog(fmt.Sprintf("HTTP Listener on %s:%d", SITE_HOST, SITE_HTTP_PORT))
		failOnErr(http.ListenAndServe(fmt.Sprintf("%s:%d", SITE_HOST, SITE_HTTP_PORT), nil), "listenerHTTP")
	} else {
		infoLog("HTTP listener disabled")
	}
}

func listenerTLS() {
	if SITE_ENABLE_TLS {
		/*config := tls.Config{
			MinVersion:               tls.VersionTLS10,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
				tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA},
		}*/
		infoLog(fmt.Sprintf("TLS Listener on %s:%d", SITE_HOST, SITE_TLS_PORT))
		failOnErr(http.ListenAndServeTLS(fmt.Sprintf("%s:%d", SITE_HOST, SITE_TLS_PORT), SITE_TLS_CERT, SITE_TLS_KEY, nil), "listenerTLS")
	} else {
		infoLog("TLS listener disabled")
	}
}
