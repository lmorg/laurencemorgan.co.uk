// post
package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const YES string = "✓"

type FormLoginRegister struct {
	Username     FormField
	Password     FormField
	Email        FormField
	Twitter      FormField
	ErrorMessage string
}

type FormField struct {
	Value string
	Error string
}

type FormComment struct {
	Post      string
	Error     string
	ParentID  uint
	CommentID uint
}

type FormThread struct {
	ThreadID uint
	Title    string
	LinkURL  string
	MIME     string
	To       string
	Error    string
	Post     string
}

func postComment(session *Session) (successful bool, f *FormComment, redirect string) {
	f = new(FormComment)

	if session.GetPost("comment").Value != YES {
		return
	}

	var (
		post   DisplayText = DisplayText{trim(session.GetPost("post").Value)}
		parent string      = session.GetPost("parent_id").Value
	)

	// TODO: check the parent_id is valid
	parent_id, _ := Atoui(parent)
	// if it is, then use it
	f.ParentID = parent_id

	if session.ID == "" {
		f.Post = post.HTMLEscaped()
		redirect = "/login/post"
		return
	}

	if !matchTokens(session, session.GetPost("token").Value) {
		f.Post = post.HTMLEscaped()
		f.Error = fmt.Sprintf("Invalid token. Please try again.\nThis is usually because the previous page had been open for more than %d hours.", CORE_TOKEN_AGE)
	}

	// check field lengths
	if utf8.RuneCountInString(post.Value) > CORE_POST_MAX_CHARS {
		f.Post = post.HTMLEscaped()
		f.Error = fmt.Sprintf("Post was longer than %d characters.", CORE_POST_MAX_CHARS)
		return
	}
	if utf8.RuneCountInString(post.Value) < CORE_POST_MIN_CHARS {
		f.Post = post.HTMLEscaped()
		f.Error = fmt.Sprintf("Post too short. Must be at least %d characters long.", CORE_POST_MIN_CHARS)
		return
	}

	// check BBCode formatting
	tag := cmsBBNoClose(&post.Value)
	if tag != "" {
		f.Post = post.HTMLEscaped()
		f.Error = fmt.Sprintf(`Mismatched '%s' tags.`, tag)
		return
	}

	// spam control
	var (
		count       byte
		thread_id   uint
		thread_type string
		meta_ref    string
		locked      string
		enabled     string
		permissions string
	)

	// check for spam and permissions
	// (i shouldn't need to check for permissions because this process shouldn't get called without sufficiant permissions)
	err := dbSelectRow(session, true, dbQueryRow(session, SQL_VALIDATE_COMMENT_POST, session.Thread.ID, f.ParentID, session.User.ID, post.Value, session.Thread.ID),
		&count, &thread_id, &thread_type, &meta_ref, &locked, &enabled, &permissions)

	if err != nil {
		f.Post = session.GetPost("post").HTMLEscaped()
		f.Error = "Unable to validate request."
		return
	}

	if count != 0 {
		f.Post = session.GetPost("post").HTMLEscaped()
		f.Error = "Duplicated post."
		return
	}

	if enabled == "N" {
		return
	}

	if locked == "Y" {
		f.Error = "This thread has been locked."
		return
	}

	// update comment table
	bbcode := post.Value
	cmsBBDecode(&bbcode, session)
	insPost := &bbcode
	cached := "Y"
	if len(bbcode) > CORE_POST_MAX_CHARS {
		debugLog("[post] [postComment] [bbcode] too long!!!!")
		cached = "N"
		insPost = &post.Value
	}
	transaction, r, err := dbInsertRow(session, true, false, SQL_INSERT_COMMENT,
		session.Page.Section.ID, *insPost, cached, parent_id, session.User.ID)

	if err != nil {
		//f.Post = post.HTMLEscaped()
		f.Error = err.Error()
		return
	}

	// update thread headers
	_, err = dbInsertTransaction(session, transaction, true, true, SQL_INSERT_COMMENT_UPDATE_THREAD,
		session.Now, session.User.ID, session.Page.Section.ID)

	if err != nil {
		//f.Post = post.HTMLEscaped()
		f.Error = err.Error()
		return
	}

	successful = true

	RealTimeUpdate(session, "thread", &bbcode)

	// get last row ID
	i, err := r.LastInsertId()
	if err != nil {
		isErr(session, err, true, "getting row ID", "postComment")
		return
	}

	// write comment history record
	_, _, err = dbInsertRow(session, true, true, SQL_INSERT_COMMENT_HISTORY,
		i, post.Value, "", session.r.RemoteAddr, trimString(session.r.UserAgent(), CORE_USER_AGENT_CROP).Value)
	if err != nil {
		isErr(session, err, true, "writing history record", "postComment")
	}

	f.CommentID = uint(i)
	session.SetCookie("cnew", Itoa(f.CommentID), 10)

	return
}

////////////////////////////////////////////////////////////////////////////////

func postThread(session *Session, thread_type, thread_model string) (successful bool, f *FormThread, cid uint) {
	//////////////////////////////////
	// FIRST WE VALIDATE THE THREAD //
	//////////////////////////////////

	f = new(FormThread)

	if session.GetPost("thread").Value != YES {
		return
	}

	var (
		//link_description string
		post    DisplayText = DisplayText{trim(session.GetPost("post").Value)}
		title   DisplayText = DisplayText{trim(session.GetPost("title").Value)}
		linkurl DisplayText = DisplayText{trim(session.GetPost("linkurl").Value)}
		mime    DisplayText = DisplayText{trim(session.GetPost("mime").Value)}
		to      DisplayText = DisplayText{trim(session.GetPost("to").Value)}
		a_to    []string
	)

	// write form data back on error
	defer func() {
		if f.Error != "" {
			f.Title = title.HTMLEscaped()
			f.LinkURL = linkurl.HTMLEscaped()
			f.Post = post.HTMLEscaped()
			f.MIME = mime.HTMLEscaped()
			f.To = to.Value
		}
	}()

	if session.ID == "" {
		s_log_reg := "login"
		if CORE_ALLOW_REGISTRATION {
			s_log_reg += " or register"
		}
		f.Error = fmt.Sprintf(`You need to <a href="%slogin" title="%s: %s">%s</a> before you can create new threads.`,
			SITE_HOME_URL, SITE_NAME, SITE_DESCRIPTION, s_log_reg)

		return
	}

	if !matchTokens(session, session.GetPost("token").Value) {
		f.Error = fmt.Sprintf("Invalid token. Please try again.\nThis is usually because the previous page had been open for more than %d hours.", CORE_TOKEN_AGE)
		return
	}

	if utf8.RuneCountInString(title.Value) > CORE_THREAD_TITLE_MAX_C {
		f.Error = fmt.Sprintf("Title too long. Must be no more than %d characters long.", CORE_THREAD_TITLE_MAX_C)
		return
	}
	if utf8.RuneCountInString(title.Value) < CORE_THREAD_TITLE_MIN_C {
		f.Error = fmt.Sprintf("Title too short. Must be at least %d characters long.", CORE_THREAD_TITLE_MIN_C)
		return
	}

	if thread_type != THREAD_TYPE_URL {
		// check field lengths

		if utf8.RuneCountInString(post.Value) > CORE_POST_MAX_CHARS {
			f.Error = fmt.Sprintf("Post was longer than %d characters.", CORE_POST_MAX_CHARS)
			return
		}
		if utf8.RuneCountInString(post.Value) < CORE_POST_MIN_CHARS {
			f.Error = fmt.Sprintf("Post too short. Must be at least %d characters long.", CORE_POST_MIN_CHARS)
			return
		}

		// check BBCode formatting
		tag := cmsBBNoClose(&post.Value)
		if tag != "" {
			f.Error = fmt.Sprintf("Mismatched '%s' tags.", tag)
			return
		}

		if thread_type == THREAD_TYPE_PM {

			if len(to.Value) < 2 {
				f.Error = "No users selected"
				return
			}

			/*if to.Value !~ m/[0-9,]+/ { //TODO: not write this in Perl
				to.Value == ""
				f.Error = "Invalid list of recipients"
				return
			}*/

			a_to = strings.Split(to.Value, ",")
		}

	} else { // thread_type == THREAD_TYPE_URL

		// check field lengths
		if linkurl.Value == "" {
			f.Error = "No link URL supplied."
			return
		}
		if utf8.RuneCountInString(post.Value) > CORE_THREAD_LINK_DESC_MAX_C {
			f.Error = fmt.Sprintf("Description was longer than %d characters.", CORE_POST_MAX_CHARS)
			return
		}
		if utf8.RuneCountInString(post.Value) < CORE_THREAD_LINK_DESC_MIN_C {
			f.Error = fmt.Sprintf("Description too short. Must be at least %d characters long.", CORE_POST_MIN_CHARS)
			return
		}

		if utf8.RuneCountInString(linkurl.Value) > CORE_THREAD_LINK_URL_MAX_C {
			f.Error = fmt.Sprintf("Link URL too long. Must be no more than %d characters long.", CORE_THREAD_LINK_URL_MAX_C)
			return
		}

		if utf8.RuneCountInString(mime.Value) > CORE_THREAD_LINK_CT_MAX_C {
			f.Error = fmt.Sprintf("Content-Type too long. Please report this error to %s administrators", SITE_NAME)
			return
		}
		if mime.Value == "" {
			f.Error = fmt.Sprintf("No Content-Type was found for %s", linkurl.HTMLEscaped())
			return
		}

		//link_description = post.Value

	}

	/*
		// TODO: investigate why this weird code exists.
		var count byte

		err := dbSelectRow(session, true, dbQueryRow(session, SQL_VALIDATE_COMMENT_POST, thread_id, session.Thread.ID), &count)
		if err != nil {
			f.Post = session.GetPost("post").Value
			f.Error = "Unable to validate request."
			return
		}

		if count != 1 {
			f.Post = session.GetPost("post").Value
			f.Error = "Cannot post comment."
			return
		}
	*/

	///////////////////////////////////////
	// NOW WRITE THIS THREAD TO DATABASE //
	///////////////////////////////////////

	// insert thread
	transaction, r, err := dbInsertRow(session, true, false, SQL_INSERT_THREAD,
		session.Forum.ID, title.Value, linkurl.Value, mime.Value, post.Value, /*link_description*/
		thread_type, session.User.ID, session.User.ID, thread_model,
		session.r.RemoteAddr, trimString(session.r.UserAgent(), CORE_USER_AGENT_CROP).Value)

	defer func() {
		// close transaction regardess of point of exit
		if f.Error != "" {
			if err := transaction.Rollback(); err != nil {
				f.Error += "<p>&nbsp;</p>" + err.Error()
			}

		} else {
			if err := transaction.Commit(); err != nil {
				isErr(session, err, true, "transaction.Commit()", "postThread")
				f.Error = err.Error()
			}
		}

		if f.Error == "" {
			successful = true

			// update forum cache
			f_obj := new(CacheForums)
			f_obj.Init()
			f_obj.Cache.f = append(f_obj.Cache.f, Forum{
				ForumID:     session.Forum.ID,
				UpdatedDate: session.Now,
			})
			f_obj.Cache.f[0].Latest = &f_obj.Cache.f[0].UpdatedDate
			f_obj.Method = CACHE_FORUMS_METHOD_NEW_THREAD
			cache_forums <- f_obj
		}
	}()

	if err != nil {
		//f.Post = post.HTMLEscaped()
		f.Error = err.Error()
		return
	}

	i, err := r.LastInsertId()
	if err != nil {
		isErr(session, err, true, "getting row ID (thread)", "postThread")
		return
	}

	if thread_type == THREAD_TYPE_URL {
		f.ThreadID = uint(i)
		return

	} else {
		// add post to thread
		bbcode := post.Value
		cached := "Y"
		cmsBBDecode(&bbcode, session)
		if len(bbcode) > CORE_POST_MAX_CHARS {
			debugLog("[post] [postThread] [bbcode] wow - thats some long content!!")
			bbcode = post.Value
			cached = "N"
		}
		_, err = dbInsertTransaction(session, transaction, true, false, SQL_INSERT_COMMENT,
			i, bbcode, cached, 0, session.User.ID)

		if err != nil {
			//f.Post = post.HTMLEscaped()
			f.Error = err.Error()
			return
		}

		f.ThreadID = uint(i)

		i, err = r.LastInsertId()
		if err != nil {
			f.Error = err.Error()
			isErr(session, err, true, "getting row ID (post)", "postThread")
			return
		}

		// write comment history record
		_, err = dbInsertTransaction(session, transaction, true, false, SQL_INSERT_COMMENT_HISTORY,
			i, post.Value, "", session.r.RemoteAddr, trimString(session.r.UserAgent(), CORE_USER_AGENT_CROP).Value)

		if err != nil {
			f.Error = err.Error()
			isErr(session, err, true, "writing history record", "postThread")
			return
		}

		cid = uint(i)
		session.SetCookie("cnew", Itoa(cid), 10)
	}

	if thread_type == THREAD_TYPE_PM {
		sql_insert_thread_pm := SQL_INSERT_THREAD_PM
		for _, uid := range a_to {
			if uid != "" {
				sql_insert_thread_pm += fmt.Sprintf(" (%d, %s),", f.ThreadID, uid)
			}
		}
		sql_insert_thread_pm += fmt.Sprintf(" (%d, %d)", f.ThreadID, session.User.ID)

		_, err = dbInsertTransaction(session, transaction, true, false, sql_insert_thread_pm)

		if err != nil {
			f.Error = err.Error()
			isErr(session, err, true, "inserting user list", "postThread")
			return
		}
	}

	return

}

////////////////////////////////////////////////////////////////////////////////

func postLogin(session *Session) (successful bool, f *FormLoginRegister) {
	f = new(FormLoginRegister)

	f.Username.Value = strings.TrimSpace(session.GetPost("username").Value)
	password := strings.TrimSpace(session.GetPost("password").Value)

	if f.Username.Value == "" {
		f.ErrorMessage = "No username entered."
		return
	}

	if password == "" {
		f.ErrorMessage = "No password entered."
		return
	}

	var (
		uid uint
		pw  string
	)

	err := dbSelectRow(session, true, dbQueryRow(session, SQL_VALIDATE_LOGIN, f.Username.Value), &uid, &session.User.JoinDate, &pw, &session.ID, &session.User.Hash, &session.User.Salt)

	if err != nil || uid == 0 {
		f.ErrorMessage = "Invalid username."
		return
	}

	if pw != passwordHash(session, password) {
		f.ErrorMessage = "Incorrect password."
		return
	}

	session.WriteSessionCookies(true)

	successful = true
	return
}

func postRegister(session *Session) (successful bool, f *FormLoginRegister) {
	f = new(FormLoginRegister)
	var (
		password1 string
		password2 string
	)

	f.Username.Value = strings.TrimSpace(session.GetPost("username").Value)
	password1 = strings.TrimSpace(session.GetPost("password1").Value)
	password2 = strings.TrimSpace(session.GetPost("password2").Value)
	f.Email.Value = strings.TrimSpace(session.GetPost("email").Value)
	f.Twitter.Value = strings.TrimSpace(session.GetPost("twitter").Value)

	if utf8.RuneCountInString(f.Username.Value) < CORE_USERNAME_MIN_CHARS {
		f.ErrorMessage = fmt.Sprintf("Username must be %d or more characters long!", CORE_USERNAME_MIN_CHARS)
		f.Username.Error = fmt.Sprintf("Min %d characters.", CORE_USERNAME_MIN_CHARS)
	}
	if utf8.RuneCountInString(password1) < CORE_PASSWORD_MIN_CHARS {
		f.ErrorMessage = fmt.Sprintf("Password must be %d or more characters long!<br/>A mixture of alpha and numerics is recommended, but not enforced.", CORE_PASSWORD_MIN_CHARS)
		f.Password.Error = fmt.Sprintf("Min %d characters.", CORE_PASSWORD_MIN_CHARS)
	}
	if utf8.RuneCountInString(f.Username.Value) > CORE_USERNAME_MAX_CHARS {
		f.ErrorMessage = "Username too long!"
		f.Username.Error = fmt.Sprintf("Max %d characters.", CORE_USERNAME_MAX_CHARS)
	}
	if utf8.RuneCountInString(password1) > CORE_PASSWORD_MAX_CHARS {
		f.ErrorMessage = fmt.Sprintf("Password too long! %d character limit.", CORE_PASSWORD_MAX_CHARS)
		f.Password.Error = fmt.Sprintf("Max %d characters.", CORE_PASSWORD_MAX_CHARS)
	}
	if utf8.RuneCountInString(f.Email.Value) > CORE_EMAIL_MAX_CHARS {
		f.ErrorMessage = "Email address too long!"
		f.Email.Error = fmt.Sprintf("Max %d characters.", CORE_EMAIL_MAX_CHARS)
	}
	if utf8.RuneCountInString(f.Twitter.Value) > CORE_TWITTER_MAX_CHARS {
		f.ErrorMessage = "Twitter username too long!"
		f.Twitter.Error = fmt.Sprintf("Max %d characters.", CORE_TWITTER_MAX_CHARS)
	}
	if password1 != password2 {
		f.ErrorMessage = "Passwords do not match!"
		f.Password.Error = "Must match."
	}
	//if (!Email::Valid->address($email)) {
	//    f.ErrorMessage = "Invalid email address!";
	//    f.Email.Error = "Invalid email.";
	//}
	//if ($twitter =~ m/[^a-zA-Z0-9_]/) {
	//    f.ErrorMessage = "Not a valid Twitter username!";
	//    $errors{'rerr_twitter'} = "Invalid.";
	//}

	runes := []rune(f.Username.Value)
	//for i := 0; i < utf8.RuneCountInString(f.Username.Value); i++ {
	for _, r := range runes {
		//r, _ := utf8.DecodeRuneInString(f.Username.Value[i : i+1])
		if !strconv.IsPrint(r) {
			f.ErrorMessage = "Username must only contain printable characters (inc. space)!"
			f.Username.Error = "Invalid."
			break
		}
	}

	if f.Username.Value == "" || password1 == "" || password2 == "" ||
		(f.Email.Value == "" && CORE_REG_EMAIL_REQUIRED) || (f.Twitter.Value == "" && CORE_REG_TWITTER_REQUIRED) {
		f.ErrorMessage = "One or more required fields empty!"
	}

	if f.ErrorMessage != "" {
		return
	}

	var (
		username_matches int
		email_matches    int
		err              error
	)

	err = dbSelectRow(session, true, dbQueryRow(session, SQL_VALIDATE_REGISTRATION, strings.ToLower(f.Username.Value), strings.ToLower(f.Email.Value)),
		&username_matches, &email_matches)

	if err != nil {
		f.ErrorMessage = SITE_NAME + " was unable to validate your registration."
		return
	}

	if username_matches > 0 {
		f.ErrorMessage = "Username already registered!"
		f.Username.Error = "Must be unique."
	}
	if email_matches > 0 && (f.Email.Value != "" || CORE_REG_EMAIL_REQUIRED) {
		f.ErrorMessage = "e-mail address already registered!"
		f.Email.Error = "Must be unique."
	}

	if f.ErrorMessage != "" {
		return
	}

	session.User.Name.Alias.Value = f.Username.Value
	session.User.Email.Value = f.Email.Value
	session.User.Twitter.Name = f.Twitter.Value
	err = userAdd(session, password1)
	if err != nil {
		f.ErrorMessage = "Failed to add user: " + err.Error()
		return
	}
	successful = true
	return
}
