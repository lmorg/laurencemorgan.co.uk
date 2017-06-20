// facebook
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	fb_login            = "https://www.facebook.com/dialog/oauth?client_id=%s&redirect_uri=%s&state=%s"
	fb_graphapi_token   = "https://graph.facebook.com/oauth/access_token?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s"
	fb_graphapi_accinfo = "https://graph.facebook.com/me?access_token=%s"
	fb_graphapi_avatar  = "https://graph.facebook.com/%s/picture?width=%d&height=%d&redirect=false&access_token=%s"
)

func facebookLogin(session *Session) string {
	session.w.Header().Set("Cache-Control", "no-cache")
	var (
		code           string
		attempt        string
		state          string
		err            error
		uri            string
		http_resp      *http.Response
		http_body      []byte
		json_interface interface{}
		json_values    map[string]interface{}
		user           User
	)

	// check if there's a return from facebook
	code = session.r.FormValue("code")
	attempt = session.GetCookie("attempt").Value
	// this is a bit of a cheap trick, but it saves writing a bespoke parser
	//token, err = url.Parse("/url/?" + string(http_body))
	//if err == nil {
	//	user.Facebook.AccessToken = token.Query().Get("access_token")
	//}

	// if we have had a response from facebook then lets check our validation hash
	// if we haven't, then lets create a new validation hash.
	if code != "" && session.r.URL.Query().Get("state") != "" {
		state = session.GetCookie("secret").Value
	} else {
		state = validationHash(session, "facebook login "+time.Now().Format(DATE_FMT_ARTICLE))
		session.SetCookie("secret", state, 600) // give the user 10 minutes to log in
	}

	////////////////////////////////////////////////////////////////////////////
	// error signing in (eg permissions denied)
	if session.r.URL.Query().Get("error") != "" {
		var (
			err               error
			error_reason      string
			error_description string
		)

		error_reason, err = url.QueryUnescape(session.r.URL.Query().Get("error_reason"))
		error_reason = strings.ToUpper(strings.Replace(error_reason, "_", " ", -1))
		isErr(session, err, true, "reading facebook return values", "facebookLogin")

		error_description, err = url.QueryUnescape(session.r.URL.Query().Get("error_description"))
		isErr(session, err, true, "reading facebook return values", "facebookLogin")

		facebookLoginFailed(session, "Failed to login via Facebook: "+error_description, nil)
		return ""
	}

	////////////////////////////////////////////////////////////////////////////
	// display facebook login page / permissions page
	if code == "" || session.r.URL.Query().Get("state") == "" {
		return fmt.Sprintf(fb_login, FACEBOOK_APP_ID, url.QueryEscape(SITE_HOME_URL+"login/facebook/"), state)
	}

	////////////////////////////////////////////////////////////////////////////
	// CSRF check (though I'd expect this to fail more frequently because of the 10 minute time out)
	if state != session.r.URL.Query().Get("state") {
		if attempt == "" {
			// firefox seems to do some weird caching. So lets force a second attempt then give up
			session.SetCookie("attempt", "retry", 5)
			return SITE_HOME_URL + "login/facebook/?code="
		}
		facebookLoginFailed(session, "Session keys do not match. Were you idle on the Facebook sign in page for more than 10 minutes?", nil)
		return ""
	}

	////////////////////////////////////////////////////////////////////////////
	// now lets get some data from Facebook....
	//http_resp, err = http.Get(fmt.Sprintf(fb_graphapi_token, FACEBOOK_APP_ID, url.QueryEscape(SITE_HOME_URL+"login/facebook/"), FACEBOOK_APP_SECRET, code))
	uri = fmt.Sprintf(fb_graphapi_token, FACEBOOK_APP_ID, url.QueryEscape(SITE_HOME_URL+"login/facebook/"), FACEBOOK_APP_SECRET, code)
	http_resp, err = httpRequest(httpClient(session, uri))

	if err == nil {
		http_body, err = ioutil.ReadAll(http_resp.Body)
		http_resp.Body.Close()
		if err == nil {
			//var j interface{}
			//err = json.Unmarshal(http_body, &j)
			err := json.Unmarshal(http_body, &json_interface)
			if err == nil {
				json_values = json_interface.(map[string]interface{})
				user.Facebook.AccessToken = json_values["access_token"].(string)
			}
			// this is a bit of a cheap trick, but it saves writing a bespoke parser
			//token, err = url.Parse("/url/?" + string(http_body))
			//if err == nil {
			//	user.Facebook.AccessToken = token.Query().Get("access_token")
			//}
		}
	}

	// did we receive an access token?
	if user.Facebook.AccessToken == "" {
		debugLog(string(http_body))
		facebookLoginFailed(session, "Unable to obtain an access token from facebook.", err)
		return ""
	}

	////////////////////////////////////////////////////////////////////////////
	// get json
	uri = fmt.Sprintf(fb_graphapi_accinfo, user.Facebook.AccessToken)
	http_resp, err = httpRequest(httpClient(session, uri))
	if err == nil {
		http_body, err = ioutil.ReadAll(http_resp.Body)
		//if err == nil {
		http_resp.Body.Close()
		//}
	}
	if err != nil {
		facebookLoginFailed(session, "Communication with Facebook died.", err)
		return ""
	}

	// parse the first json file
	json.Unmarshal(http_body, &json_interface)
	json_values = json_interface.(map[string]interface{})
	user.Facebook.ID = json_values["id"].(string)
	//user.Facebook.URL = json_values["link"].(string) //TODO: do i really need this since I'm not actually storing it in the DB?
	user.Name.First.Value = json_values["first_name"].(string)
	user.Name.Full.Value = json_values["name"].(string)

	////////////////////////////////////////////////////////////////////////////
	// get profile picture
	//http_resp, err = http.Get(fmt.Sprintf(fb_graphapi_avatar, user.Facebook.ID, CORE_AVATAR_WIDTH, CORE_AVATAR_HEIGHT, user.Facebook.AccessToken))
	uri = fmt.Sprintf(fb_graphapi_avatar, user.Facebook.ID, CORE_AVATAR_LARGE_WIDTH, CORE_AVATAR_LARGE_HEIGHT, user.Facebook.AccessToken)
	http_resp, err = httpRequest(httpClient(session, uri))
	if err == nil {
		http_body, err = ioutil.ReadAll(http_resp.Body)
		if err == nil {
			http_resp.Body.Close()
		}
	}
	if err != nil {
		facebookLoginFailed(session, "Communication with Facebook died.", err)
		return ""
	}

	// parse second json feed
	json.Unmarshal(http_body, &json_interface)
	json_values = json_interface.(map[string]interface{})
	//debugLog(json_values)

	var profile_picture_url string
	if json_values["data"].(map[string]interface{})["is_silhouette"].(bool) == false {
		profile_picture_url = json_values["data"].(map[string]interface{})["url"].(string)
		user.AvatarValue = "Y"
	} else {
		user.AvatarValue = "N"
	}

	//debugLog(user)
	err = facebookDB(session, user)
	if err != nil {
		return ""
	}

	if user.AvatarValue == "Y" {
		//httpGetBinary(session, profile_picture_url, fmt.Sprintf("%savatars/%d-large.jpg", PWD_WRITABLE_PATH, session.User.ID))
		if response, err := httpRequest(httpClient(session, profile_picture_url)); err == nil {
			resizeAvatar(session, response.Body)
		}
	}
	return getReferrer(session)
}

func facebookLoginFailed(session *Session, message string, err error) {
	session.Page.Content = fmt.Sprintf(`<h1>Something went wrong 😞</h1><p>%s</p>
		<p>Would you like to <a href="%slogin/facebook/?code=" alt="retry logging into %s with Facebook">retry logging in with Facebook</a>,
		or <a href="%slogin/" alt="Login to %s">sign in with another account</a>?</p>`, message, SITE_HOME_URL, SITE_NAME, SITE_HOME_URL, SITE_NAME)
	isErr(session, err, false, "facebook login", "facebookLogin")
	session.SetCookie("secret", "", -1)
	session.SetCookie("attempt", "", -1)
}

func facebookDB(session *Session, user User) (err error) {
	var (
		count      int
		user_id    interface{}
		session_id interface{}
		user_hash  interface{}
		salt       interface{}
	)
	// check if facebook user already exists
	err = dbSelectRow(session, false, dbQueryRow(session, SQL_VALIDATE_FACEBOOK, user.Facebook.ID),
		&count, &user_id, &user_hash, &session_id, &salt)

	if err != nil {
		facebookLoginFailed(session, SITE_NAME+" failed to look you up on the user database.", err)
		return
	}

	// if not, create a new user
	if count == 0 {
		session.User = user
		err = userAdd(session, "")
		facebookLoginFailed(session, SITE_NAME+" failed to add you to the user database.", err)
		return
	}

	// multiple accounts
	if count > 1 {
		err = fmt.Errorf("Multiple facebook accounts: user id = %d", user_id.(int64))
		facebookLoginFailed(session, "You seem to have multiple records for this Facebook account on our database. This should never happen.", err)
	}

	// log in
	if count == 1 {
		session.User.Hash = string(user_hash.([]byte))
		if string(session_id.([]byte)) != "" {
			session.ID = string(session_id.([]byte))
		} else {
			_, _, err = dbInsertRow(session, true, true, SQL_LOGIN_FACEBOOK, session.CreateSessionID(), user.Facebook.ID, user.Facebook.AccessToken, user.AvatarValue)
		}
		session.WriteSessionCookies(true)
		session.User.ID = uint(user_id.(int64))
		session.User.Salt = string(salt.([]byte))
		return
	}

	err = fmt.Errorf("Unknown error")
	facebookLoginFailed(session, "Login routine bombed out.", err)
	return
}
