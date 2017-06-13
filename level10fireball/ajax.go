// ajax.go
package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func pageAJAX(session *Session) bool {
	session.w.Header().Set("Cache-Control", "max-age=0")
	if strings.Contains(session.r.Header.Get("Accept-Encoding"), "gzip") {
		session.w.Header().Set("Content-Encoding", "gzip")
	}
	session.w.Header().Set("Content-Type", "text/plain")
	if fnAJAX[session.Path[2]] == nil {
		setStatus(session, http.StatusNotFound)
		session.Page.Content = "{404}"

	} else {
		fnAJAX[session.Path[2]](session)
	}

	writeBody(session)

	return true
}

var fnAJAX = map[string]func(*Session){

	"quoteComment": func(session *Session) {
		var (
			content    string
			err        error
			comment_id uint
		)

		comment_id, _ = Atoui(session.Path[3])

		err = dbSelectRow(session, true,
			dbQueryRow(session, SQL_SELECT_LATEST_COMMENT_BBCODE, comment_id),
			&content)

		if err == nil {
			session.Page.Content = "[quote]\n" + content + "\n[/quote]\n"
		}

	},

	"threadNextPage": func(session *Session) {
		// get thread id
		session.Thread.ID, _ = Atoui(session.Path[3])
		if session.Thread.ID < 1 {
			return
		}

		// get page number
		session.Page.nPageCurrent, _ = Atoui(session.Path[4])
		if session.Page.nPageCurrent < 1 {
			session.Page.nPageCurrent = 1
		}

		var (
			err               error
			thread_locked     string
			thread_type       string
			thread_model      string
			s_rc              interface{}
			forum_id          interface{}
			forum_title       interface{}
			forum_desc        interface{}
			article_id        interface{}
			article_title     interface{}
			article_desc      interface{}
			topic_id          interface{}
			topic_title       interface{}
			topic_desc        interface{}
			link_url          string
			link_content_type string
			link_desc         string
		)

		if session.Path[2] != "pm" {
			// normal threads
			err = dbSelectRow(session, true, dbQueryRow(session, SQL_THREAD_HEADERS, session.User.ID, session.Thread.ID, session.User.Permissions.RegEx()),
				&session.Thread.Title.Value, &thread_locked, &thread_type, &thread_model, &s_rc,
				&forum_id, &forum_title, &forum_desc,
				&article_id, &article_title, &article_desc,
				&topic_id, &topic_title, &topic_desc,
				&link_url, &link_content_type, &link_desc)
		} else {
			err = dbSelectRow(session, true, dbQueryRow(session, SQL_PM_HEADERS, session.Thread.ID, session.User.ID, session.Thread.ID, session.User.ID),
				&session.Thread.Title.Value, &thread_locked, &thread_type, &thread_model, &s_rc,
				&forum_id, &forum_title, &forum_desc,
				&article_id, &article_title, &article_desc,
				&topic_id, &topic_title, &topic_desc,
				&link_url, &link_content_type, &link_desc)
		}

		// on error, silent fail
		if err != nil {
			return
		}

		session.ReadComments = session.ReadComments.Make(s_rc)
		forumViewThread(session, thread_model, 0, 0)
		session.ReadComments.Export(session)

		return
	},

	"giveKarma": func(session *Session) {
		comment_id, _ := Atoui(session.GetQueryString("c").Value)

		if session.User.ID == 0 || comment_id == 0 {
			session.Page.Content = ""
			return
		}

		var (
			has_permissions interface{}
			no_banned_karma interface{}
		)

		if !matchTokens(session, session.GetQueryString("t").Value) {
			session.Page.Content = "Invalid token. Please refresh this page"
			return
		}

		err := dbSelectRow(session, true, dbQueryRow(session, SQL_VALIDATE_COMMENT_KARMA,
			comment_id, comment_id, session.User.ID, session.User.Permissions.RegEx(), session.User.Permissions.RegEx()),
			&has_permissions, &no_banned_karma)

		// on error, silent fail
		if err != nil || has_permissions == nil {
			session.Page.Content = "denied"
			return
		}

		if no_banned_karma != nil && no_banned_karma.([]uint8)[0] == 'N' {
			session.Page.Content = "You have banned karma against this comment"
			return
		}

		// insert an audit trail
		_, _, err = dbInsertRow(session, true, true, SQL_UPDATE_COMMENT_KARMA,
			comment_id, session.User.ID, 1, "", time.Now(), "Y")
		if err != nil {
			session.Page.Content = "failed"
			return
		}

		// sum the audit trail, ready to update the comment record
		var sum int
		err = dbSelectRow(session, true, dbQueryRow(session, SQL_SUM_COMMENT_KARMA, comment_id), &sum)
		if err != nil {
			session.Page.Content = "added"
			return
		}

		// now update the comment record
		_, _, err = dbInsertRow(session, true, true, SQL_UPDATE_COMMENT_SUMMED_KARMA, sum, comment_id)

		if err != nil {
			session.Page.Content = "added"
			return
		}

		// return new karma
		session.Page.Content = strconv.Itoa(sum)
		return
	},

	"urlPreview": func(session *Session) {
		var (
			err       error
			url_parse URLParse
			//client    *http.Client
			//request   *http.Request
			response *http.Response
		)

		if session.User.ID == 0 || !matchTokens(session, session.GetQueryString("t").Value) {
			json_encoded_error := errors.New("Invalid or expired token.<br/>Please refresh this page and try again")
			url_parse := URLParse{Err: json_encoded_error}
			session.Page.Content, err = url_parse.ToJSONobj()
			if err != nil {
				session.w.Header().Set("error", err.Error())
			}
			return
		}

		// create http client - return a failure if err
		response, url_parse.Err = httpRequest(httpClient(session, session.GetQueryString("u").Value))
		if url_parse.Err != nil {
			session.Page.Content, err = url_parse.ToJSONobj()
			if err != nil {
				session.w.Header().Set("error", err.Error())
			}

			return

		}

		// GET URL
		session.Page.Content, err = ogGet(response).ToJSONobj()
		if err != nil {
			session.w.Header().Set("error", err.Error())
		}

		return
	},

	"userList": func(session *Session) {
		if session.User.ID != 0 && matchTokens(session, session.GetQueryString("t").Value) {
			session.Page.Content = cache_users.JSON[session.Path[3]]
		}
		return
	},

	"commentDetails": func(session *Session) {
		if session.User.ID == 0 || !matchTokens(session, session.GetQueryString("t").Value) &&
			!session.User.Permissions.Match(PERM_FORUM_MOD+PERM_FORUM_ADMIN+PERM_ROOT) {

			session.Page.Content = "Permission granted"

		}
		return
	},
}
