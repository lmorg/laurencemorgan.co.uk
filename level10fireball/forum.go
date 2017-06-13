// forum
package main

import (
	"fmt"
	"html"
	"github.com/lmorg/laurencemorgan.co.uk/level10fireball/fmtd"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	// 0 == use global default
	COMMENTS_NONE       byte = 1
	COMMENTS_HIGHLIGHTS byte = 2
	COMMENTS_ALL        byte = 3

	THREAD_TYPE_ARTICLE string = "A"
	THREAD_TYPE_FORUM   string = "F"
	THREAD_TYPE_PM      string = "P"
	THREAD_TYPE_URL     string = "U"

	THREAD_MODEL_FLAT     string = "flat"
	THREAD_MODEL_THREADED string = "threaded"
)

type Forum struct {
	ForumID       uint
	ParentID      uint
	CreatedDate   time.Time
	UpdatedDate   time.Time
	Latest        *time.Time
	UpdatedStr    string
	Title         string
	Description   string
	ThreadCount   uint
	ReadPerm      string // read permission
	NewThreadPerm string // new thread permission
	ThreadType    string // default thread type for new threads
	ThreadModel   string // default thread model for new threads
	//CreatedUser User
	//UpdatedUser User
}

type Thread struct {
	URL           string
	ID            uint
	Title         string
	Model         string
	Type          string
	_createdStr   string
	CreatedTime   time.Time
	CreatedUser   Name
	_updatedStr   string
	UpdatedTime   time.Time
	UpdatedUser   Name
	_latestStr    string
	LatestTime    *time.Time
	LatestUser    *Name
	_lastVisit    interface{}
	LastVisitTime time.Time
	_subscribed   interface{}
	CommentsN     int
	_readComments interface{}
}

func CacheThread(thread *Thread) {
	threadType := func(s string) string {
		if s == THREAD_TYPE_PM {
			return "pm"
		}
		return "thread"
	}

	thread.URL = fmt.Sprintf("%s%s/%d/%s", SITE_HOME_URL, threadType(thread.Type), thread.ID, urlify(thread.Title))

	thread.CreatedTime = dateParse(thread._createdStr, "CacheThread")
	if thread._updatedStr[:4] == "0000" {
		thread.LatestTime = &thread.CreatedTime
		thread.LatestUser = &thread.CreatedUser
	} else {
		thread.UpdatedTime = dateParse(thread._updatedStr, "CacheThread")
		thread.LatestTime = &thread.UpdatedTime
		thread.LatestUser = &thread.UpdatedUser
	}
}

///////////////////////////////////////////////////////////////////////////////

type Comment struct {
	ThreadID    uint
	ID          uint
	ParentID    uint
	User        User
	Content     string
	Cached      string
	Created     string
	Updated     interface{}
	LatestDate  string
	Karma       uint
	Rank        int64
	ThreadModel *string
}

type ReadComments struct {
	ByID    map[uint]bool
	Highest uint
}

func (rc ReadComments) Make(text interface{}) ReadComments {
	rc.ByID = make(map[uint]bool)
	rc.Highest = 0

	if text == nil {
		return rc
	}

	var (
		s  []string
		id uint
	)
	s = strings.Split(string(text.([]byte)), "|")
	for i := 0; i < len(s); i++ {
		id, _ = Atoui(s[i])
		if id != 0 {
			if id > rc.Highest {
				rc.Highest = id
			}
			rc.ByID[id] = true
		}
	}

	return rc
}

func (rc ReadComments) Export(session *Session) {
	var (
		id   uint
		text string
	)
	for id, _ = range rc.ByID {
		text += "|" + Itoa(id)
	}
	for id, _ = range session.UnreadComments {
		text += "|" + Itoa(id)
	}

	if session.User.ID > 0 {
		dbInsertRow(session, false, true, SQL_UPDATE_THREAD_VIEWED,
			session.Thread.ID, session.User.ID, session.Now, text)
	}
}

func forumReplyForm(session *Session, thread_locked string, f *FormComment, thread_type string) {
	if thread_locked == "Y" {
		session.Page.Content += `<span id="postcomment"><p>This thread has been locked.</p></span>`
		return
	}

	var (
		error_message string
		parent_id     uint
		post_url      string
	)
	if f != nil {
		error_message = f.Error
		parent_id = f.ParentID
	}

	if thread_type != THREAD_TYPE_PM {
		post_url = fmt.Sprintf("%sthread/%d/%s#postcomment", SITE_HOME_URL, session.Thread.ID, session.Page.Section.Title.URLify())
	} else {
		post_url = fmt.Sprintf("%spm/%d/%s#postcomment", SITE_HOME_URL, session.Thread.ID, session.Page.Section.Title.URLify())
	}

	session.Page.Content += fmt.Sprintf(`
			<div id="reply">
                <div id="reply_header">
    				<h2><span id="postcomment">Leave a comment</span></h2>
    				<div id="reply_quote" style="display:none">
    					<_quote|1>
    					    <span id="reply_to_content"></span>
    					<_quote>
    				</div>
    				<span id="post-error" class="auth_err">%s</span>
    			</div>
    			<div id="reply_form" class="blur">
    		        <form action="%s" method="post" id="form_reply_form" accept-charset="UTF-8">
    		            <textarea name="post" rows="5" class="reply" maxlength="%d" inputmode="user predicton"
    		                onkeydown="autoGrow(this);textCounter()"
    		                onkeyup="autoGrow(this);textCounter();postToCache(%d)"
        		            onfocus="delClass('reply_form','blur');addClass('reply_form','focus');"
		   					onblur="delClass('reply_form','focus');addClass('reply_form','blur');postToCache(%d);"
                        >%s</textarea>
    					<input type="hidden" name="parent_id" value="%d" />
    		            <input type="hidden" name="token" value="%s" />
    		            
    		            <span class="comment_helpers nowrap comment_clear"><a href="javascript:postClear(%d);" title="Clear editor and cached copy.">Clear editor</a></span>

    		            <button name="comment" value="%s" type="submit" class="btn right">Post comment!</button>
    		            <span id="forum_char_limit" class="nowrap comment_helpers right">%d%s</span><span class="comment_helpers nowrap ondesktop right"><a href="javascript:showHiddenContainer('forum_bbcode_instructions');" title="View supported BBCode.">BBCode supported</a></span>
                    </form>
                </div>

    		    <div id="forum_bbcode_instructions" class="onmobile">
                    %s
    		    </div>
            </div>`,
		error_message,                        // post error
		post_url,                             // form action
		CORE_POST_MAX_CHARS,                  // maxlength="%d"
		session.Thread.ID, session.Thread.ID, // postToCache(%d)
		session.GetPost("post").HTMLEscaped(), //post
		parent_id, session.Token, //             input type="hidden"
		session.Thread.ID,                                      // post clear
		YES,                                                    // submit button
		CORE_POST_MAX_CHARS, lang(session, "char_limit").Value, // character count
		session.layout.BBCode)

	var post_cache string
	if session.GetCookie("cnew").Value == "" {
		post_cache = "postFromCache"
	} else {
		post_cache = "postRmCache"
	}
	session.PostProcInc += fmt.Sprintf(`<script>var char_limit_str='%s',char_limit_max=%d,pid=id('form_reply_form').elements['parent_id'].value;if(pid>0&&id("cir_%d").innerHTML!='Quote')replyTo(pid);%s(%d);autoGrow(id('form_reply_form').elements['post']);textCounter();</script>`, lang(session, "char_counter").JSEscaped(), CORE_POST_MAX_CHARS, parent_id, post_cache, session.Thread.ID)
}

func forumNewThread(session *Session, thread_type string, f *FormThread) {
	switch thread_type {
	case THREAD_TYPE_URL:
		forumNewURLThread(session, f)
		return
	case THREAD_TYPE_PM:
		forumNewPMThread(session, f)
		return
	}

	var (
		error_message string
	)

	if f != nil {
		error_message = f.Error
	}

	session.Page.Content += fmt.Sprintf(`
			<div id="reply">
                <div id="reply_header">
    				<h2><span id="createthread">Create a new thread</span></h2>
    				<span id="post-error" class="auth_err">%s</span>
    			</div>
    		    <form action="%sforum/%d/%s#createthread" method="post" id="form_reply_form" accept-charset="UTF-8">
                    <div class="nt_title reg_required">Title</div><div class="nt_field"><input type="text" name="title" class="textbox w100" value="%s" maxlength="%d" /></div>
                    <div class="clear"></div>
                    <div id="reply_form" class="blur">
        		        <textarea name="post" rows="5" class="reply" maxlength="%d" inputmode="user predicton"
        		            onkeydown="autoGrow(this);textCounter();"
        		            onkeyup="autoGrow(this);textCounter();"
        		            onfocus="delClass('reply_form','blur');addClass('reply_form','focus');"
		   					onblur="delClass('reply_form','focus');addClass('reply_form','blur');"
        		        >%s</textarea>
        		        <input type="hidden" name="token" value="%s" />
        		        <button name="thread" value="%s" type="submit" class="btn right">Create thread!</button>
    		            <span id="forum_char_limit" class="nowrap comment_helpers right">%d%s</span><span class="comment_helpers nowrap ondesktop right"><a href="javascript:showHiddenContainer('forum_bbcode_instructions');" title="View supported BBCode.">BBCode supported</a></span>
                    </div>
                </form>

    		    <div id="forum_bbcode_instructions" class="onmobile">
                    %s
    		    </div>
            </div>`,
		error_message,                                                 // post error
		SITE_HOME_URL, session.Forum.ID, session.Forum.Title.URLify(), // form action
		session.GetPost("title").HTMLEscaped(), CORE_THREAD_TITLE_MAX_C, // thread title
		CORE_POST_MAX_CHARS, session.GetPost("post").HTMLEscaped(), // thread post
		session.Token,                                          // hidden
		YES,                                                    // submit
		CORE_POST_MAX_CHARS, lang(session, "char_limit").Value, // character count
		session.layout.BBCode)

	session.PostProcInc += fmt.Sprintf(`<script>var char_limit_str='%s',char_limit_max=%d;autoGrow(id('form_reply_form').elements['post']);textCounter();</script>`, lang(session, "char_counter").JSEscaped(), CORE_POST_MAX_CHARS)
}

func forumNewURLThread(session *Session, f *FormThread) {
	var (
		error_message string
	)

	if f != nil {
		error_message = f.Error
	}

	session.Page.Content += fmt.Sprintf(`
			<div id="reply">
                <div id="reply_header">
    				<h2><span id="submiturl">Submit new link</span></h2>
    				<span id="post-error" class="auth_err">%s</span>
    			</div>
    		    <form action="%sforum/%d/%s#submiturl" method="post" id="form_reply_form" accept-charset="UTF-8">
                    <div class="clear"></div>
                    <div class="nt_title reg_required">Article Source (URL)</div>
                    	<div class="nt_field"><input type="text" name="linkurl" class="textbox w100" value="%s" maxlength="%d" onblur="urlPreview();" placeholder="http://example.com/..."/></div>
                    	<div class="clear"></div>
                    <div id="parse_err_container" class="show hide">
                        <div id="parse_err" class="error"></div>
                        <div class="clear"></div>
                    </div>
                    <div class="nt_title reg_required">Title</div><div class="nt_field"><input type="text" name="title" class="textbox w100" value="%s" maxlength="%d"/></div>
                    <div class="clear"></div>
                    <div id="reply_form" class="blur">
        		        <textarea name="post" rows="2" class="reply" maxlength="%d" inputmode="user predicton"
        		            onkeydown="autoGrow(this);textCounter();"
        		            onkeyup="autoGrow(this);textCounter();"
        		            onfocus="delClass('reply_form','blur');addClass('reply_form','focus');"
		   					onblur="delClass('reply_form','focus');addClass('reply_form','blur');"
        		        >%s</textarea>
        		        <input type="hidden" name="mime" value="%s" />
        		        <input type="hidden" name="token" value="%s" />
        		        <button name="thread" value="%s" type="submit" class="btn right">Submit link!</button>
                        <span id="forum_char_limit" class="nowrap comment_helpers right">%d%s</span>
                    </div>
                </form>
            </div>`,
		error_message,                                                 //        post error
		SITE_HOME_URL, session.Forum.ID, session.Forum.Title.URLify(), //        form action
		session.GetPost("linkurl").HTMLEscaped(), CORE_THREAD_LINK_URL_MAX_C, // thread URL
		session.GetPost("title").HTMLEscaped(), CORE_THREAD_TITLE_MAX_C, //      thread title
		CORE_THREAD_LINK_DESC_MAX_C, session.GetPost("post").HTMLEscaped(), //   thread post
		session.GetPost("mime").HTMLEscaped(), session.Token, //                 hidden fields
		YES, //                                                                  submit
		CORE_THREAD_LINK_DESC_MAX_C, lang(session, "char_limit").Value) //       character count

	session.PostProcInc += fmt.Sprintf(`<script type="text/javascript">var char_limit_str='%s',char_limit_max=%d;autoGrow(id('form_reply_form').elements['post']);textCounter();</script>`,
		lang(session, "char_counter").JSEscaped(), CORE_THREAD_LINK_DESC_MAX_C)
}

func forumNewPMThread(session *Session, f *FormThread) {
	var (
		error_message string
	)

	if f != nil {
		error_message = f.Error
	}

	session.Page.Content += fmt.Sprintf(`
			<div id="reply">
                <div id="reply_header">
    				<h2><span id="pvtconv">Start a private conversation</span></h2>
    				<span id="post-error" class="auth_err">%s</span>
    			</div>
    		    <form action="%spm/folder/%d/%s#pvtconv" method="post" id="form_reply_form" accept-charset="UTF-8">
                    <div class="nt_title reg_required">To</div>
                    	<div class="nt_field">%s</div>
                    	<div class="clear"></div>
                    <div class="nt_title reg_required">Title</div>
                    	<div class="nt_field"><input type="text" name="title" class="textbox w100" value="%s" maxlength="%d" /></div>
                    	<div class="clear"></div>
                    <div id="reply_form" class="blur">
        		        <textarea name="post" rows="5" class="reply" maxlength="%d" inputmode="user predicton"
        		            onkeydown="autoGrow(this);textCounter();"
        		            onkeyup="autoGrow(this);textCounter();"
        		            onfocus="delClass('reply_form','blur');addClass('reply_form','focus');"
		   					onblur="delClass('reply_form','focus');addClass('reply_form','blur');"
		   				>%s</textarea>
						<input type="hidden" id="ul_form_to" name="to" value="%s" />
        		        <input type="hidden" name="token" value="%s" />
        		        <button name="thread" value="%s" type="submit" class="btn right">Send PM!</button>
    		            <span id="forum_char_limit" class="nowrap comment_helpers right">%d%s</span><span class="comment_helpers nowrap ondesktop right"><a href="javascript:showHiddenContainer('forum_bbcode_instructions');" title="View supported BBCode.">BBCode supported</a></span>
                    </div>
                </form>

    		    <div id="forum_bbcode_instructions" class="onmobile">
                    %s
    		    </div>
            </div>`,
		error_message,                                                 // post error
		SITE_HOME_URL, session.Forum.ID, session.Forum.Title.URLify(), // form action
		session.layout.SelectUserTemplate,                               // user list dropdown
		session.GetPost("title").HTMLEscaped(), CORE_THREAD_TITLE_MAX_C, // thread title
		CORE_POST_MAX_CHARS, session.GetPost("post").HTMLEscaped(), // thread post
		session.GetPost("to").HTMLEscaped(), session.Token, // hidden fields
		YES,                                                    // submit button
		CORE_POST_MAX_CHARS, lang(session, "char_limit").Value, // character count
		session.layout.BBCode)

	session.PostProcInc += fmt.Sprintf(`<script>var char_limit_str='%s',char_limit_max=%d;autoGrow(id('form_reply_form').elements['post']);textCounter();</script>`, lang(session, "char_counter").JSEscaped(), CORE_POST_MAX_CHARS)
}

func forumHighlights(session *Session) (n_comments uint, thread_locked string, html *string) {
	var (
		errp          error
		comment       Comment
		comments_html string
		i             uint
		thread_model  string
	)

	comments := dbQueryRows(session, false, &errp, SQL_SHOW_COMMENT_HIGHLIGHTS,
		session.Thread.ID, session.Thread.ID, session.layout.nCommentsPerHighlight)

	// pretty out of db sessions error message
	if dbOutOfSessions(session, errp) {
		return
	}

	for i < session.layout.nCommentsPerHighlight && dbSelectRows(session, false, &errp, comments,
		&thread_locked, &thread_model, &comment.ID, &comment.ParentID,
		&comment.User.ID, &comment.User.Name.Alias.Value, &comment.User.Name.First.Value, &comment.User.Name.Full.Value, &comment.User.AvatarValue,
		&comment.Content, &comment.Cached, &comment.Created, &comment.Updated, &comment.Karma, &n_comments) {

		// pretty out of db sessions error message
		if dbOutOfSessions(session, errp) {
			return
		}

		comment.ThreadID = session.Thread.ID
		comment.ThreadModel = &thread_model

		if session.Mobile {
			return n_comments, thread_locked, &comments_html
		}

		comments_html += *forumComment(session, &comment, THREAD_MODEL_FLAT, 0)
		i++
	}

	return n_comments, thread_locked, &comments_html
}

func forumComment(session *Session, comment *Comment, model string, cnew uint) *string {
	var (
		comment_html  string
		created       time.Time
		updated       time.Time
		update_msg    string
		latest        *time.Time
		reply_caption string
		me            string
	)

	// quote or reply
	if *comment.ThreadModel == THREAD_MODEL_FLAT {
		reply_caption = "Quote"
	} else {
		reply_caption = "Reply"
	}

	// comment models
	if comment.User.ID == session.User.ID {
		model += " c_uid_me"
		me = "me"
	} else if session.ReadComments.ByID[comment.ID] {
		model += " c_uid_other"
	} else {
		model += " c_uid_new"
		session.UnreadComments[comment.ID] = true
		//debugLog("unread comment", comment.ID)
	}

	// have i just posted it?
	if comment.ID == cnew {
		model += " fadein"
	}

	// get the latest date
	created = dateParse(comment.Created, "forumComment")
	if comment.Updated != nil {
		u := string(comment.Updated.([]uint8))
		if u[:4] != "0000" {
			updated = dateParse(u, "forumComment")
			update_msg = fmt.Sprintf(`<br/>Edited: %s`, updated.Format(DATE_FMT_THREAD))
			latest = &updated
		} else {
			latest = &created
		}
	} else {
		latest = &created
	}

	if comment.Cached != "Y" {
		cmsBBDecode(&comment.Content, session)
	}

	comment_html = strings.Replace(session.layout.CommentTemplate, "<_comment|id>", Itoa(comment.ID), -1)
	comment_html = strings.Replace(comment_html, "<_comment|username>", comment.User.Name.Long().HTMLEscaped(), -1)
	comment_html = strings.Replace(comment_html, "<_comment|user id>", Itoa(comment.User.ID), -1)
	comment_html = strings.Replace(comment_html, "<_comment|me>", me, -1)
	comment_html = strings.Replace(comment_html, "<_comment|content>", comment.Content, -1)
	comment_html = strings.Replace(comment_html, "<_comment|avatar|small>", comment.User.AvatarSmall(), -1)
	comment_html = strings.Replace(comment_html, "<_comment|avatar|large>", comment.User.AvatarLarge(), -1)
	comment_html = strings.Replace(comment_html, "<_comment|karma>", Itoa(comment.Karma), -1)
	comment_html = strings.Replace(comment_html, "<_comment|created>", created.Format(DATE_FMT_THREAD), -1)
	comment_html = strings.Replace(comment_html, "<_comment|updated>", update_msg, -1)
	comment_html = strings.Replace(comment_html, "<_comment|latest>", latest.Format(DATE_FMT_THREAD), -1)
	comment_html = strings.Replace(comment_html, "<_comment|age>", fmt.Sprintf("%v", session.Now.Sub(*latest)), -1)

	comment_html = strings.Replace(comment_html, "<_comment|reply>", reply_caption, -1)
	comment_html = strings.Replace(comment_html, "<_comment|permalink>", fmt.Sprintf("%sc/%d", rx_html_prefix.ReplaceAllString(SITE_HOME_URL, ""), comment.ID), -1)
	comment_html = strings.Replace(comment_html, "<_comment|model>", model, -1)

	return &comment_html
}

func forumAllForumThreads(session *Session, thread *Thread, placement uint, threads_html *string) {
	var (
		new_comments   string
		subscribed     string
		pagination_str string
	)

	if session.User.ID > 0 {
		if thread._lastVisit != nil {
			thread.LastVisitTime = dateParse(string(thread._lastVisit.([]uint8)), "forumAllForumThreads")
			subscribed = string(thread._subscribed.([]uint8))
			if thread.LastVisitTime.Before(*thread.LatestTime) {
				new_comments = "new"
			} else {
				new_comments = ""
			}
		} else {
			new_comments = "new"
			subscribed = "N"
		}
	} else {
		new_comments = ""
		subscribed = "N"
	}

	if thread.Model == THREAD_MODEL_FLAT {
		pagination_str = paginationCustom(session, thread.CommentsN, session.layout.nCommentsFlatPerPage, thread.URL+"/page")
		if thread._readComments != nil {
			session.ReadComments = session.ReadComments.Make(thread._readComments)
			thread.URL += fmt.Sprintf("?id=%d#comment%d", session.ReadComments.Highest, session.ReadComments.Highest)
		}
	} else {
		pagination_str = ""
	}

	thread_item := strings.Replace(session.layout.ForumsTemplateThread, "<_thread|id>", Itoa(thread.ID), -1)
	thread_item = strings.Replace(thread_item, "<_thread|title>", html.EscapeString(thread.Title), -1)
	thread_item = strings.Replace(thread_item, "<_thread|created|date>", thread.CreatedTime.Format(DATE_FMT_THREAD), -1)
	thread_item = strings.Replace(thread_item, "<_thread|created|alias>", thread.CreatedUser.Short().HTMLEscaped(), -1)
	thread_item = strings.Replace(thread_item, "<_thread|created|name>", thread.CreatedUser.Long().HTMLEscaped(), -1)
	thread_item = strings.Replace(thread_item, "<_thread|updated|date>", thread.UpdatedTime.Format(DATE_FMT_THREAD), -1)
	thread_item = strings.Replace(thread_item, "<_thread|updated|alias>", thread.UpdatedUser.Short().HTMLEscaped(), -1)
	thread_item = strings.Replace(thread_item, "<_thread|updated|name>", thread.UpdatedUser.Long().HTMLEscaped(), -1)
	thread_item = strings.Replace(thread_item, "<_thread|latest|date>", thread.LatestTime.Format(DATE_FMT_THREAD), -1)
	thread_item = strings.Replace(thread_item, "<_thread|latest|alias>", thread.LatestUser.Short().HTMLEscaped(), -1)
	thread_item = strings.Replace(thread_item, "<_thread|latest|name>", thread.LatestUser.Long().HTMLEscaped(), -1)
	thread_item = strings.Replace(thread_item, "<_thread|age>", fmt.Sprintf("%v", fmtd.Fuzzy(session.Now.Sub(*thread.LatestTime))), -1)
	thread_item = strings.Replace(thread_item, "<_thread|new>", new_comments, -1)
	thread_item = strings.Replace(thread_item, "<_thread|subscribed>", subscribed, -1)
	thread_item = strings.Replace(thread_item, "<_thread|comments>", strconv.Itoa(thread.CommentsN), -1)
	thread_item = strings.Replace(thread_item, "<_thread|url>", thread.URL, -1)
	thread_item = strings.Replace(thread_item, "<_thread|placement>", Itoa(placement), -1)
	thread_item = strings.Replace(thread_item, "<_thread|pagination>", pagination_str, -1)
	*threads_html += thread_item
}

func forumViewThread(session *Session, model string, parent_id uint, cnew uint) {
	if model == "" {
		return
	}
	fnThreadModels[model](session, parent_id, cnew)
}

var fnThreadModels = map[string]func(*Session, uint, uint){

	THREAD_MODEL_FLAT: func(session *Session, parent_id uint, cnew uint) {

		// starting page if parent_id specified
		if parent_id != 0 {
			var comment_num uint
			err := dbSelectRow(session, true, dbQueryRow(session, SQL_SHOW_COMMENTS_FLAT_BY_PID, session.Thread.ID, parent_id), &comment_num)
			if dbOutOfSessions(session, err) {
				return
			}

			session.Page.nPageCurrent = uint(math.Ceil(float64(comment_num) / float64(session.layout.nCommentsFlatPerPage)))
		}

		var errp error

		rows := dbQueryRows(session, false, &errp, SQL_SHOW_COMMENTS_FLAT,
			session.Thread.ID /*session.User.ID,*/, session.Thread.ID, session.User.Permissions.RegEx(),
			(session.Page.nPageCurrent-1)*session.layout.nCommentsFlatPerPage, session.layout.nCommentsFlatPerPage)

		// pretty out of db sessions error message
		if dbOutOfSessions(session, errp) {
			return
		}

		var (
			c             Comment
			c_plain       string
			thread_title  string
			thread_locked string
			thread_type   string
			thread_model  string = "flat"
			//karma_id      interface{}
		)

		session.Page.Content += fmt.Sprintf(`<h3>Page %d:</h3>`, session.Page.nPageCurrent)

		for dbSelectRows(session, false, &errp, rows,
			&thread_title, &thread_locked, &thread_type, &c.ID, &c.ParentID,
			&c.User.ID, &c.User.Name.Alias.Value, &c.User.Name.First.Value, &c.User.Name.Full.Value, &c.User.AvatarValue,
			&c.Content, &c.Cached, &c.Created, &c.Updated, &c.Karma,
			&session.Page.nPageCount) {

			c.ThreadModel = &thread_model

			if c.Cached == "Y" {
				c_plain = trimString(cmsPlainFormatting(&c.Content), 50).Value
			} else {
				c_plain = trimString(cmsPurgeBBCode(&c.Content), 50).HTMLEscaped()
			}

			session.Page.Content += fmt.Sprintf(`<div id="nesth%d" class="c_unhide hidden0">
                                                    <a href="javascript:showBranch(%d);">unhide comment</a><br/><b>%s:</b> %s</div>
                                                 <div id="nest%d" class="show">`,
				c.ID, c.ID,
				c.User.Name.Short().HTMLEscaped(),
				c_plain,
				c.ID)
			session.Page.Content += *forumComment(session, &c, THREAD_MODEL_FLAT, cnew)
			session.Page.Content += "</div>"
		}

		// TODO: only for flat threads
		session.Page.nPageCount = uint(math.Ceil(float64(session.Page.nPageCount) / float64(session.layout.nCommentsFlatPerPage)))
		session.Page.Content += fmt.Sprintf(`<script type="text/javascript">max_pages=%d</script><div id="t_page%d" class="t_page">`, session.Page.nPageCount, session.Page.nPageCurrent)
		if session.Path[1] != "ajax" {
			session.Page.Content += pagination(session, SITE_HOME_URL+session.Path[1]+"/"+session.Path[2]+"/"+session.Path[3]+`/page%d`)
		} else {
			session.Page.Content += pagination(session, SITE_HOME_URL+session.Path[5]+"/"+session.Path[3]+"/"+session.Path[6]+`/page%d`)
		}
		session.Page.Content += `</div><div class="clear"></div>`
	},

	////////////////////////////////////////////////////////////////////////////
	THREAD_MODEL_THREADED: func(session *Session, parent_id uint, cnew uint) {
		var errp error

		rows := dbQueryRows(session, false, &errp, SQL_SHOW_COMMENTS_THREADED, session.Thread.ID, parent_id, parent_id, session.User.Permissions.RegEx())

		// pretty out of db sessions error message
		if dbOutOfSessions(session, errp) {
			return
		}

		var (
			comments      map[uint][]Comment = make(map[uint][]Comment)
			c             Comment
			thread_title  string
			thread_locked string
			thread_type   string
			thread_model  string = THREAD_MODEL_THREADED
		)

		for dbSelectRows(session, false, &errp, rows,
			&thread_title, &thread_locked, &thread_type, &c.ID, &c.ParentID,
			&c.User.ID, &c.User.Name.Alias.Value, &c.User.Name.First.Value, &c.User.Name.Full.Value, &c.User.AvatarValue,
			&c.Content, &c.Cached, &c.Created, &c.Updated, &c.Karma, &c.Rank) {

			c.ThreadModel = &thread_model
			comments[c.ParentID] = append(comments[c.ParentID], c)

			if parent_id == c.ID {
				session.Page.Content += fmt.Sprintf(`<div class="c_less_prev">↖ Comment from <a href="%s#comment%d" title="%s">previous page</a>:</div>`,
					getReferrer(session), c.ParentID, session.Page.Section.Title.HTMLEscaped()) // TODO: this is a terrible implimentation. fix so it doesn't use cookies!
				session.Page.Content += *forumComment(session, &c, "threaded1", cnew)
				session.Page.Content += `<div class="c_less_cont">↘ Discussion continued:</div>`
			}
		}

		forumViewThreadThreadedNest(session, &comments, parent_id, 1, cnew)
	},
}

func forumViewThreadThreadedNest(session *Session, comments *map[uint][]Comment, parent_id uint, level uint, cnew uint) {
	for i, _ := range (*comments)[parent_id] {
		if level > CORE_MAX_COMMENT_NEST {
			session.Page.Content += fmt.Sprintf(`<div class="c_more"><a href="%s?id=%d" title="Continue reading this branch">↪ continue reading this branch</a></div>`,
				session.r.URL.Path, parent_id)
			return
		} else {

			var c_plain string

			if (*comments)[parent_id][i].Cached == "Y" {
				c_plain = trimString(cmsPlainFormatting(&(*comments)[parent_id][i].Content), 50).Value
			} else {
				c_plain = trimString(cmsPurgeBBCode(&(*comments)[parent_id][i].Content), 50).HTMLEscaped()
			}

			session.Page.Content += fmt.Sprintf(`<div id="nesth%d" class="c_unhide hidden0">
                                                    <a href="javascript:showBranch(%d);">unhide branch</a><br/><b>%s:</b> %s</div>
                                                 <div id="nest%d" class="show">`,
				(*comments)[parent_id][i].ID, (*comments)[parent_id][i].ID,
				(*comments)[parent_id][i].User.Name.Short().HTMLEscaped(),
				c_plain,
				(*comments)[parent_id][i].ID)

			session.Page.Content += *forumComment(session, (&(*comments)[parent_id][i]), "threaded"+Itoa(level), cnew)
			forumViewThreadThreadedNest(session, comments, (*comments)[parent_id][i].ID, level+1, cnew)

			session.Page.Content += `</div><div class="thread_end"></div>`
		}
	}
}
