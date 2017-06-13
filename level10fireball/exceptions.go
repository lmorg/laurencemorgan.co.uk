// exeptions
package main

import (
	"errors"
	"fmt"
	"log"
	//	_ "net/http/pprof" // TODO: delete if not DEBUG
	"os"
	"strconv"
	"time"
)

const DEBUG bool = true

func accessLog(session *Session) {
	isEmpty := func(s string) string {
		if s == "" {
			return "-"
		}
		return s
	}

	isZero := func(i int) string {
		if i == 0 {
			return "-"
		}
		return strconv.Itoa(i)
	}

	// TODO: redirect access log somewhere
	fmt.Printf("%s %s %s [%s] \"%s %s %s\" %s %s \"%s\" \"%s\" %d %s\n",
		session.r.RemoteAddr, isZero(int(session.User.ID)), "-",
		session.Now.Format("02/Jan/2006:15:04:05 -0700"),
		session.r.Method, session.r.RequestURI, session.r.Proto,
		isZero(session.Status), isZero(session.ResponseSize),
		isEmpty(session.r.Referer()), session.r.UserAgent(),
		time.Now().Sub(session.Now).Nanoseconds()/1000,
		session.Theme,
	)
}

func debugLog(args ...interface{}) {
	// TODO: redirect debug log somewhere
	if DEBUG {
		args = append([]interface{}{"[debug]"}, args...)
		log.Println(args...)
	}
}

func infoLog(args ...interface{}) {
	// TODO: redirect error log somewhere
	args = append([]interface{}{"[info]"}, args...)
	log.Println(args...)
}

func errLog(args ...interface{}) {
	// TODO: redirect error log somewhere
	args = append([]interface{}{"[error]"}, args...)
	log.Println(args...)
}

////////////////////////////////////////////////////////////////////////////////

func raiseErr(session *Session, msg string, surpress_error_msg bool, friendly string, function string) (err error) {
	err = errors.New(msg)
	isErr(session, err, surpress_error_msg, friendly, function)
	return err
}

// dies if error caught
func failOnErr(err error, function_name string) {
	if err != nil {
		//errLog(err)
		log.Printf("[fail] [%s] [%s]", function_name, err)
		os.Exit(1)
	}
}

func isErr(session *Session, err error, surpressErrorMsg bool, friendly string, function string) error {
	if err != nil {
		msg := errWriteHTML(session, err, friendly, function)
		if !surpressErrorMsg {
			session.Page.Content += msg
		}
	}
	return err
}

func errWriteHTML(session *Session, err error, friendly string, function string) (s string) {
	var (
		ip      string = "-"
		user_id string = "-"
		//uri     string = "-"
	)

	if err != nil {
		s = fmt.Sprintf(`<div class="error">Error in %s (function: %s):<br/>%s</div>`, friendly, function, err.Error())
		if session != nil && session.r != nil {
			ip = session.r.RemoteAddr
			user_id = Itoa(session.User.ID)
		}
		//log.Printf("%sError in %s (function: %s): %s\n", ip, friendly, function, err.Error())
		//errLog(ip, "Error in", friendly, "(function:", function, "): ", err.Error())
		errLog(fmt.Sprintf("[IP:%s] [UserID:%s] [Func:%s] [Friendly:%s] %s", ip, user_id, function, friendly, err.Error()))
	}
	return
}
