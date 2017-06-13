// layout
package main

import (
	"database/sql"
	"fmt"
	"html"
	"level10fireball/fmtd"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Layout struct {
	Quotes                   []string
	Desktop                  PageLayout
	Mobile                   PageLayout
	SocialButtons            string
	PrivacyTemplate          string
	BBCode                   string
	Menubar                  []Menubar
	MenubarLeft              string
	MenubarRight             string
	BreadcrumbSeparator      string
	ArticleTemplateBC        string // before comments
	ArticleTemplateAC        string // after comments
	ArticlesTemplate         string
	ArticlesTemplateArticle  string
	ArticleAppendComments    byte
	TopicsTemplate           string
	TopicsTemplateTopic      string
	TopicsTemplateArticle    string
	PaginationTemplate       string
	ThreadTemplateBC         string // before comments
	ThreadTemplateAC         string // after comments
	ForumsTemplate           string
	ForumsTemplateForum      string
	ForumsTemplateSubForum   string
	ForumsTemplateThread     string
	NewThreadTemplate        string
	ShowCommentTemplate      string // page
	CommentTemplate          string // item
	ShowUserTemplate         string
	ShowMeTemplate           string
	SelectUserTemplate       string
	GalleryTemplates         map[string]GalleryTemplate
	nListColumns             uint // typically the number of columns for thumbnails
	nForumRowsColours        uint // typically the number of shades for the forum table
	nItemsPerPage            uint // number of thumbnails per page
	nArticlesPerTopic        uint // number of articles to show per thumbnail
	nCommentsPerHighlight    uint // number of comments to show in the highlights
	nCommentNestsPerPage     uint // number of comment indentations per page
	nCommentsThreadedPerPage uint // number of top level comments per page when in threaded view
	nCommentsFlatPerPage     uint // number of comments per page when in list view
	nCharsMobileBreadcrumbs  uint // max number of characters per item on mobile breadcrumbs
	xEmbeddedFrame           uint
	yEmbeddedFrame           uint
}

type Menubar struct {
	Label       string
	Description string
	URL         string
}

type PageTemplate struct {
	Header string
	Footer string
}

type PageSession struct {
	PageTemplate
	Content      string
	Images       []string
	nPageCurrent uint
	nPageCount   uint
	Section      *Section
}

type PageLayout struct {
	PageTemplate
	//	RealTime bool
}

func (page PageSession) Title() DisplayText {
	return page.Section.Title
}

func (page PageSession) Description() DisplayText {
	return page.Section.Description
}

type Section struct {
	Title       DisplayText
	Description DisplayText
	ID          uint
}

func (section Section) GetURL(section_name string) string {
	return fmt.Sprintf("%s%s/%d/%s", SITE_HOME_URL, section_name, section.ID, section.Title.URLify())
}

func (section Section) GetShortURL(section_name string) string {
	return fmt.Sprintf("%s%s/%d/%s", rx_html_prefix.ReplaceAllString(SITE_HOME_URL, ""), section_name[:1], section.ID, section.Title.URLify())
}

type DisplayText struct {
	Value string
}

func (dt DisplayText) HTMLEscaped() string {
	return html.EscapeString(dt.Value)
}

func (dt DisplayText) QueryEscaped() string {
	return url.QueryEscape(dt.Value)
}

func (dt DisplayText) URLify() string {
	return urlify(dt.Value)
}

func (dt DisplayText) JSEscaped() string {
	return strings.Replace(dt.Value, `'`, `\'`, -1)
}

type GalleryTemplate struct {
	Gallery string
	Thumb   string
}

var (
	live_layout Layout

	rx_page_number *regexp.Regexp
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())

	rx_page_number = regexCompile(`(?i)^page[0-9]+$`)
}

func paginationCustom(session *Session, max int, increment uint, url string) (s string) {

	//<a href="`+url+`" class="selected">%d</a>
	n_pages := int(math.Ceil(float64(max) / float64(increment)))

	for i := 1; i <= 3 && i <= n_pages; i++ {
		s += fmt.Sprintf(`<a href="%s%d" class="unselected">%d</a>`, url, i, i)
	}
	for i := n_pages - 3; i <= n_pages; i++ {
		if i <= 3 {
			continue
		}
		s += fmt.Sprintf(`<a href="%s%d" class="unselected">%d</a>`, url, i, i)
	}
	return
}

func pagination(session *Session, url string) (s string) {
	s = strings.Replace(session.layout.PaginationTemplate, "<_pagination|current>", Itoa(session.Page.nPageCurrent), -1)
	s = strings.Replace(s, "<_pagination|count>", Itoa(session.Page.nPageCount), -1)
	s = strings.Replace(s, "<_pagination|smart links>", paginationSmartLinks(session, url), -1)
	return
}

func paginationSmartLinks(session *Session, url string) (s string) {
	var nbsp string
	if session.Page.nPageCurrent > 2 {
		s = fmt.Sprintf(`<a href="`+url+`" class="unselected">1</a> ...`, 1)
	}
	for i := session.Page.nPageCurrent - 1; i <= session.Page.nPageCurrent+1; i++ {
		if i > 1 {
			nbsp = "&nbsp;"
		}
		if i > 0 && i <= session.Page.nPageCount {
			if i == session.Page.nPageCurrent {
				s += fmt.Sprintf(`%s<a href="`+url+`" class="selected">%d</a>`, nbsp, i, i)
			} else {
				s += fmt.Sprintf(`%s<a href="`+url+`" class="unselected">%d</a>`, nbsp, i, i)
			}

		}
	}
	if session.Page.nPageCurrent+1 < session.Page.nPageCount {
		s += fmt.Sprintf(` ... <a href="`+url+`" class="unselected">%d</a>`, session.Page.nPageCount, session.Page.nPageCount)
	}

	return
}

func pageHome(session *Session) {
	dbSelectRow(session, false, dbQueryRow(session, SQL_LAYOUT, "home page"), &session.Page.Content)

	session.Page.Section = &session.Special
	session.Special.Title.Value = SITE_DESCRIPTION
	session.Special.Description.Value = SITE_DESCRIPTION

}

func pageArticle(session *Session) {
	var (
		str_created     string
		article_created time.Time
		date_message    string
		str_updated     string
		article_updated time.Time
	)
	session.Article.ID, _ = Atoui(session.Path[2])

	session.Page.Section = &session.Article

	// load article
	err := dbSelectRow(session, true, dbQueryRow(session, SQL_SHOW_ARTICLE, session.Article.ID, session.User.Permissions.RegEx(), session.User.Permissions.RegEx()),
		&session.Article.Title.Value, &session.Page.Content, &str_created, &str_updated,
		&session.Topic.ID, &session.Topic.Title.Value, &session.Topic.Description.Value, &session.Thread.ID)

	// pretty out of db sessions error message
	if dbOutOfSessions(session, err) {
		return
	}

	// 404 if no article found
	if session.Page.Content == "" {
		session.Article.Title.Value = "Article not found"
		page404(session, "article")
		return
	}

	// work out latest date
	if str_updated[:4] == "0000" {
		article_created = dateParse(str_created, "pageArticle")
		date_message = "Created on " + article_created.Format(DATE_FMT_ARTICLE)
	} else {
		article_updated = dateParse(str_updated, "pageArticle")
		date_message = "Last updated on " + article_updated.Format(DATE_FMT_ARTICLE)
	}

	// set the title and description
	session.Article.Description.Value = cmsPlainFormatting(&session.Page.Content)
	trimPString(&session.Article.Description.Value, 600)

	// parse the layout before loading content (stops editor abuse of layout tags)
	article_html := session.layout.ArticleTemplateBC
	cmsLayout(&article_html, session)

	// output article content
	session.Page.Content = strings.Replace(article_html, "<_content>", session.Page.Content+`<p id="articledate">`+date_message+"</p>", 1)

	// TODO: at some point I'll also make this a per article option as well
	if session.layout.ArticleAppendComments == COMMENTS_NONE {
		session.Page.Content += "Comments have not been enabled for this article."
	} else {

		// load comments
		n_comments, thread_locked, comments_html := forumHighlights(session)

		if session.Thread.ID > 0 {

			// I want this to apear above /and/ below the comments
			var comment_summary string
			if n_comments == 0 {
				comment_summary = "<p>There are no comments attached to this article. Why not be the first?</p>"
			} else {
				comment_summary = fmt.Sprintf("<p>There are %d comment%s attached to this article. ", n_comments, appendS(n_comments))
				if n_comments > session.layout.nCommentsPerHighlight || session.Mobile == true {
					comment_summary += fmt.Sprintf(`<a href="%sthread/%d/%s" title="Comments: %s">Read them all</a>.</p>`,
						SITE_HOME_URL, session.Thread.ID, urlify(session.Article.Title.Value), urlify(session.Article.Title.Value))
				}
			}

			// now lets display the comments
			session.Page.Content += `
				<hr/>
		        <div id="embedded_comments">
		        <h1>Comments</h1>`
			session.Page.Content += comment_summary
			if !session.Mobile {
				if n_comments > session.layout.nCommentsPerHighlight {
					session.Page.Content += fmt.Sprintf("<h2>The top %d comments:</h2>", session.layout.nCommentsPerHighlight)
				}
				session.Page.Content += *comments_html
				if n_comments > session.layout.nCommentsPerHighlight {
					session.Page.Content += comment_summary
				}
			}

			// reply form
			forumReplyForm(session, thread_locked, nil, THREAD_TYPE_ARTICLE)

			session.Page.Content += "</div>"
		} else {
			raiseErr(session, "Error: no forum thread created for this article", true, "displaying comment highlights", "pageArticle")
		}

	}
	// end of template
	session.Page.Content += session.layout.ArticleTemplateAC
}

func pageListArticles(session *Session) {
	var (
		errp                error
		n_items             uint
		first_item          uint
		articles            *sql.Rows
		rows_returned       bool
		placement           uint
		all_thumbs          string
		topic_id            uint
		topic_title         string
		topic_description   string
		article_id          uint
		article_title       string
		article_description string
		article_html        string
	)
	session.Page.Section = &session.Special

	// get topic ID (if set)
	topic_id, _ = Atoui(session.Path[2])

	// get page number (if not set, default to page 1)
	session.Page.nPageCurrent, _ = Atoui(session.r.URL.Query().Get("page"))
	if session.Page.nPageCurrent < 1 {
		session.Page.nPageCurrent = 1
	}

	// get how many items per page (if not set, default to CMS)
	n_items, _ = Atoui(session.r.URL.Query().Get("n"))
	if n_items < 2 {
		n_items = session.layout.nItemsPerPage
	}
	first_item = (session.Page.nPageCurrent - 1) * n_items

	if topic_id > 0 {
		articles = dbQueryRows(session, false, &errp, SQL_LIST_ALL_ARTICLES_BY_TOPIC, topic_id, session.User.Permissions.RegEx(),
			topic_id, session.User.Permissions.RegEx(), session.User.Permissions.RegEx(),
			first_item, n_items)

	} else {
		articles = dbQueryRows(session, false, &errp, SQL_LIST_ALL_ARTICLES,
			session.User.Permissions.RegEx(), session.User.Permissions.RegEx(), session.User.Permissions.RegEx(), session.User.Permissions.RegEx(),
			first_item, n_items)

		session.Special.Title.Value = "List all articles"
		session.Special.Description.Value = SITE_DESCRIPTION
	}

	// pretty out of db sessions error message
	if dbOutOfSessions(session, errp) {
		return
	}

	for dbSelectRows(session, false, &errp, articles,
		&article_id, &topic_id, &article_title, &article_description, &topic_title, &topic_description, &session.Page.nPageCount) {
		rows_returned = true
		placement++
		if placement > session.layout.nListColumns {
			placement = 1
		}

		article_description = cmsPlainFormatting(&article_description)
		trimPString(&article_description, 600)

		article_html = strings.Replace(session.layout.ArticlesTemplateArticle, "<_article|id>", Itoa(article_id), -1)
		article_html = strings.Replace(article_html, "<_article|title>", html.EscapeString(article_title), -1)
		article_html = strings.Replace(article_html, "<_article|description>", html.EscapeString(article_description), -1)
		article_html = strings.Replace(article_html, "<_article|url>", fmt.Sprintf("%sarticle/%d/%s", SITE_HOME_URL, article_id, urlify(article_title)), -1)

		article_html = strings.Replace(article_html, "<_topic|id>", Itoa(topic_id), -1)
		article_html = strings.Replace(article_html, "<_topic|title>", html.EscapeString(topic_title), -1)
		article_html = strings.Replace(article_html, "<_topic|description>", html.EscapeString(cmsPlainFormatting(&topic_description)), -1)
		article_html = strings.Replace(article_html, "<_topic|url>", fmt.Sprintf("%slist/%d/%s", SITE_HOME_URL, topic_id, urlify(topic_title)), -1)

		all_thumbs += strings.Replace(article_html, "<_article|placement>", Itoa(placement), -1)
	}

	// pretty out of db sessions error message
	if dbOutOfSessions(session, errp) {
		return
	}

	if !rows_returned {
		session.Special.Title.Value = "No articles found"
		session.Page.Content = "No articles found"
		page404(session, "")
		return
	}

	session.Page.nPageCount = uint(math.Ceil(float64(session.Page.nPageCount) / float64(n_items)))

	if session.Special.Title.Value == "" {
		session.Topic.ID = topic_id
		session.Topic.Title.Value = topic_title
		session.Topic.Description.Value = topic_description
		session.Special = session.Topic
	}

	session.Page.Content = strings.Replace(session.layout.ArticlesTemplate, "<_articles>", all_thumbs, 1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_pagination>", pagination(session, SITE_HOME_URL+session.r.URL.Path[1:]+`?page=%d`), 1)
}

func pageAllTopics(session *Session) {
	var (
		errp                error
		all_thumbs          string
		all_articles        string
		columns             uint
		topic_html          string
		topic_id            uint
		topic_title         string
		topic_description   string
		article_html        string
		article_id          uint
		article_title       string
		article_description string
	)
	session.Page.Section = &session.Special

	session.Special.Title.Value = "Topics"
	session.Special.Description.Value = SITE_DESCRIPTION

	// step through the topics
	topics := dbQueryRows(session, false, &errp, SQL_SHOW_ALL_TOPICS, session.User.Permissions.RegEx())
	// pretty out of db sessions error message
	if dbOutOfSessions(session, errp) {
		return
	}

	for dbSelectRows(session, false, &errp, topics, &topic_id, &topic_title, &topic_description) {
		all_articles = ""
		columns++
		if columns > session.layout.nListColumns {
			columns = 1
		}

		topic_html = strings.Replace(session.layout.TopicsTemplateTopic, "<_topic|id>", Itoa(topic_id), -1)
		topic_html = strings.Replace(topic_html, "<_topic|title>", html.EscapeString(topic_title), -1)
		topic_html = strings.Replace(topic_html, "<_topic|description>", html.EscapeString(cmsPlainFormatting(&topic_description)), -1)
		topic_html = strings.Replace(topic_html, "<_topic|url>", fmt.Sprintf("%slist/%d/%s", SITE_HOME_URL, topic_id, urlify(topic_title)), -1)
		topic_html = strings.Replace(topic_html, "<_topic|placement>", Itoa(columns), -1)

		// step through the articles
		articles := dbQueryRows(session, false, &errp, SQL_SHOW_ALL_TOPICS_ARTICLES, topic_id, session.User.Permissions.RegEx(), session.layout.nArticlesPerTopic)

		// pretty out of db sessions error message
		if dbOutOfSessions(session, errp) {
			return
		}

		for dbSelectRows(session, false, &errp, articles, &article_id, &article_title, &article_description) {
			article_description = cmsPlainFormatting(&article_description)
			trimPString(&article_description, 600)

			article_html = strings.Replace(session.layout.TopicsTemplateArticle, "<_article|id>", Itoa(article_id), -1)
			article_html = strings.Replace(article_html, "<_article|title>", html.EscapeString(article_title), -1)
			article_html = strings.Replace(article_html, "<_article|description>", html.EscapeString(article_description), -1)
			all_articles += strings.Replace(article_html, "<_article|url>", fmt.Sprintf("%sarticle/%d/%s", SITE_HOME_URL, article_id, urlify(article_title)), -1)
		}

		all_thumbs += strings.Replace(topic_html, "<_articles>", all_articles, -1)
	}

	// pretty out of db sessions error message
	if dbOutOfSessions(session, errp) {
		return
	}

	session.Page.Content = strings.Replace(session.layout.TopicsTemplate, "<_topics>", all_thumbs, 1)
}

func pageAllForums(session *Session) string {
	var (
		errp           error
		forum_item     string
		forums_html    string
		subforum_item  string
		subforums_html string
		placement      uint
	)

	// get forum ID (if set)
	session.Forum.ID, _ = Atoui(session.Path[2])

	session.Page.Section = &session.Forum

	cf := new(CacheForums)
	cf.Init()
	cf.Get()

	if cf.Cache.Forums[session.Forum.ID] == nil {
		session.Page.Section.Title.Value = "Forum not found"
		page404(session, "forum")
		return ""
	}

	session.Forum.Title.Value = cf.Cache.Forums[session.Forum.ID].Title
	session.Forum.Description.Value = cf.Cache.Forums[session.Forum.ID].Title + ": " + cf.Cache.Forums[session.Forum.ID].Description

	if !session.User.Permissions.Match(cf.Cache.Forums[session.Forum.ID].ReadPerm) {
		pageDenied(session, "sub-forum")
		return ""
	}

	// post a thread?
	success, form, cid := postThread(session, cf.Cache.Forums[session.Forum.ID].ThreadType, cf.Cache.Forums[session.Forum.ID].ThreadModel)
	if success {
		if cf.Cache.Forums[session.Forum.ID].ThreadType == THREAD_TYPE_PM {
			return fmt.Sprintf("%spm/%d/%s#comment%d", SITE_HOME_URL, form.ThreadID, html.EscapeString(form.Title), cid)
		} else {
			return fmt.Sprintf("%sthread/%d/%s#comment%d", SITE_HOME_URL, form.ThreadID, html.EscapeString(form.Title), cid)
		}
	}

	for _, f := range cf.Cache.Parent[session.Forum.ID] {
		// dont display forum if user doesn't have read permissions
		if !session.User.Permissions.Match(f.ReadPerm) {
			continue
		}

		if f.ForumID == 1 && CORE_HIDE_PM_FORUMS == true {
			continue
		}

		forum_item = strings.Replace(session.layout.ForumsTemplateForum, "<_forum|id>", Itoa(f.ForumID), -1)
		forum_item = strings.Replace(forum_item, "<_forum|title>", html.EscapeString(f.Title), -1)
		forum_item = strings.Replace(forum_item, "<_forum|description>", html.EscapeString(cmsPlainFormattingS(f.Description)), -1)
		forum_item = strings.Replace(forum_item, "<_forum|url>", fmt.Sprintf("%sforum/%d/%s", SITE_HOME_URL, f.ForumID, urlify(f.Title)), -1)
		var updated_str, age_str string
		if session.Mobile {
			updated_str = strings.Replace(f.UpdatedStr, " @ ", "<br/>@ ", -1)
		} else {
			updated_str = f.UpdatedStr
		}
		if updated_str != "" {
			age_str = fmt.Sprintf("%v", fmtd.Fuzzy(session.Now.Sub(f.UpdatedDate)))
		}

		forum_item = strings.Replace(forum_item, "<_forum|latest>", updated_str, -1)
		forum_item = strings.Replace(forum_item, "<_forum|age>", age_str, -1)
		forum_item = strings.Replace(forum_item, "<_forum|thread count>", fmt.Sprintf("%d thread%s", f.ThreadCount, appendS(f.ThreadCount)), -1)

		for _, subf := range cf.Cache.Parent[f.ForumID] {
			// dont display forum if user doesn't have read permissions
			if !session.User.Permissions.Match(subf.ReadPerm) {
				continue
			}
			var age_str_nested string
			if subf.UpdatedStr != "" {
				age_str_nested = fmt.Sprintf("%v", fmtd.Fuzzy(session.Now.Sub(subf.UpdatedDate)))
			}
			subforum_item = strings.Replace(session.layout.ForumsTemplateSubForum, "<_subforum|id>", Itoa(subf.ForumID), -1)
			subforum_item = strings.Replace(subforum_item, "<_subforum|title>", html.EscapeString(subf.Title), -1)
			subforum_item = strings.Replace(subforum_item, "<_subforum|description>", html.EscapeString(cmsPlainFormattingS(subf.Description)), -1)
			subforum_item = strings.Replace(subforum_item, "<_subforum|url>", fmt.Sprintf("%sforum/%d/%s", SITE_HOME_URL, subf.ForumID, urlify(subf.Title)), -1)
			subforum_item = strings.Replace(subforum_item, "<_subforum|latest>", subf.UpdatedStr, -1)
			subforum_item = strings.Replace(subforum_item, "<_subforum|age>", age_str_nested, -1)
			subforum_item = strings.Replace(subforum_item, "<_subforum|thread count>", fmt.Sprintf("%d thread%s", subf.ThreadCount, appendS(subf.ThreadCount)), -1)
			subforums_html += subforum_item
		}
		forum_item = strings.Replace(forum_item, "<_subforums>", subforums_html, 1)
		forums_html += forum_item
		subforums_html = ""
	}

	var (
		count         int
		thread        Thread
		threads_html  string
		article_title interface{}
	)

	if cf.Cache.Forums[session.Forum.ID].ThreadType != THREAD_TYPE_PM {
		// step through the threads
		rows := dbQueryRows(session, false, &errp, SQL_SHOW_ALL_FORUM_THREADS,
			session.User.ID, session.Forum.ID, session.User.Permissions.RegEx())

		// pretty out of db sessions error message
		if dbOutOfSessions(session, errp) {
			return ""
		}

		for dbSelectRows(session, false, &errp, rows,
			&thread.ID, &thread.Title, &thread._createdStr, &thread._updatedStr, &thread._latestStr, &thread._lastVisit, &thread._readComments,
			&thread.CreatedUser.Alias.Value, &thread.CreatedUser.First.Value, &thread.CreatedUser.Full.Value,
			&thread.UpdatedUser.Alias.Value, &thread.UpdatedUser.First.Value, &thread.UpdatedUser.Full.Value,
			&thread._subscribed, &thread.CommentsN, &thread.Model, &thread.Type, &article_title) {

			if article_title != nil {
				thread.Title = string(article_title.([]uint8))
			}

			placement++
			if placement > session.layout.nForumRowsColours {
				placement = 1
			}

			CacheThread(&thread)
			forumAllForumThreads(session, &thread, placement, &threads_html)

			count++
		}

	} else { // PRIVATE MESSAGES
		// step through the threads
		rows := dbQueryRows(session, false, &errp, SQL_SHOW_ALL_FORUM_PMS,
			session.User.ID, session.Forum.ID, session.User.Permissions.RegEx(), session.User.ID)

		// pretty out of db sessions error message
		if dbOutOfSessions(session, errp) {
			return ""
		}

		for dbSelectRows(session, false, &errp, rows,
			&thread.ID, &thread.Title, &thread._createdStr, &thread._updatedStr, &thread._latestStr, &thread._lastVisit, &thread._readComments,
			&thread.CreatedUser.Alias.Value, &thread.CreatedUser.First.Value, &thread.CreatedUser.Full.Value,
			&thread.UpdatedUser.Alias.Value, &thread.UpdatedUser.First.Value, &thread.UpdatedUser.Full.Value,
			&thread._subscribed, &thread.CommentsN, &thread.Model, &thread.Type) {

			if article_title != nil {
				thread.Title = string(article_title.([]uint8))
			}

			placement++
			if placement > session.layout.nForumRowsColours {
				placement = 1
			}

			CacheThread(&thread)
			forumAllForumThreads(session, &thread, placement, &threads_html)

			count++
		}
	}

	// pretty out of db sessions error message
	if dbOutOfSessions(session, errp) {
		return ""
	}

	// set the title and description depending on whether top level forum
	if session.Forum.ID != 0 {
		session.Page.Section = &session.Forum
		session.Forum.Title.Value = cf.Cache.Forums[session.Forum.ID].Title
		session.Forum.Description.Value = cf.Cache.Forums[session.Forum.ID].Title + ": " + cf.Cache.Forums[session.Forum.ID].Description
	} else {
		//if session.Forum.ID != 0 {
		session.Page.Section = &session.Special
		session.Special.Title.Value = "Forums"
		session.Special.Description.Value = SITE_DESCRIPTION
	}

	// if no sub items, check if valid forum. 404 if not, display empty forum if valid
	if session.Forum.ID != 0 && count == 0 && len(cf.Cache.Parent[session.Forum.ID]) == 0 {
		//if cf.Cache.Forums[session.Forum.ID].ForumID == 0 {
		//	session.Page.Section.Title.Value = "Forum not found"
		//	page404(session, "forum")
		//	return ""
		//} else {
		session.Page.Content = fmt.Sprintf(`<h1>The '%s' forum is currently empty 😞</h1><p>Why not be the first to start a conversation?</p>`, cf.Cache.Forums[session.Forum.ID].Title)
		//}
	} else {

		// forum not empty, so lets display the content
		session.Page.Content += strings.Replace(session.layout.ForumsTemplate, "<_forums>", forums_html, 1)
		session.Page.Content = strings.Replace(session.Page.Content, "<_threads>", threads_html, 1)

		// TODO: eh!??
		//if len(forums.c[child_id]) == 0 {
		//	session.Page.Content = strings.Replace(session.Page.Content, "<_subforums|class>", "hidden", 1)
		//}
		if count == 0 {
			session.Page.Content = strings.Replace(session.Page.Content, "<_threads|class>", "hidden", 1)
		}
	}

	if session.Forum.ID != 0 && session.ID != "" && session.User.Permissions.Match(cf.Cache.Forums[session.Forum.ID].NewThreadPerm) {
		session.Page.Content += `<hr/>`
		forumNewThread(session, cf.Cache.Forums[session.Forum.ID].ThreadType, form)
	}

	return ""
}

func pagePMs(session *Session) string {
	//session.Page.Section = &session.Thread

	// there's some nasty URL rewriting going on here
	// to kludge the use of previous code

	if session.Path[2] == "folder" {
		session.Path[2] = session.Path[3]
		return pageAllForums(session)
	}

	if session.Path[2] != "" {
		return pageThread(session)
	}

	session.Path[2] = "1" // A bit of a nasty kludge
	return pageAllForums(session)
}

func pageThread(session *Session) string {
	var (
		parent_id uint
	)
	session.Page.Section = &session.Thread

	// get thread id
	session.Thread.ID, _ = Atoui(session.Path[2])

	if session.Thread.ID < 1 {
		session.Thread.Title.Value = "Thread not found"
		page404(session, "thread")
		return ""
	}

	// get parent id (if not set, default to 0)
	parent_id, _ = Atoui(session.r.URL.Query().Get("id"))
	if parent_id < 0 {
		parent_id = 0
	}

	// get page number
	if rx_page_number.MatchString(session.Path[4]) {
		session.Page.nPageCurrent, _ = Atoui(session.Path[4][4:])
	}
	if session.Page.nPageCurrent < 1 {
		session.Page.nPageCurrent = 1
	}

	var (
		err               error
		thread_locked     string
		thread_type       string
		thread_model      string
		article_preview   string = "0"
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

	if session.Path[1] != URL_PM {
		// normal threads
		err = dbSelectRow(session, true, dbQueryRow(session, SQL_THREAD_HEADERS, session.User.ID, session.Thread.ID, session.User.Permissions.RegEx()),
			&session.Thread.Title.Value, &thread_locked, &thread_type, &thread_model, &s_rc,
			&forum_id, &forum_title, &forum_desc,
			&article_id, &article_title, &article_desc,
			&topic_id, &topic_title, &topic_desc,
			&link_url, &link_content_type, &link_desc)
	} else {
		// pm's
		err = dbSelectRow(session, true, dbQueryRow(session, SQL_PM_HEADERS, session.Thread.ID, session.User.ID, session.Thread.ID, session.User.ID),
			&session.Thread.Title.Value, &thread_locked, &thread_type, &thread_model, &s_rc,
			&forum_id, &forum_title, &forum_desc,
			&article_id, &article_title, &article_desc,
			&topic_id, &topic_title, &topic_desc,
			&link_url, &link_content_type, &link_desc)
	}

	// pretty out of db sessions error message
	if dbOutOfSessions(session, err) {
		return ""
	}
	if err != nil ||
		(thread_type != THREAD_TYPE_PM && session.Path[1] == URL_PM) ||
		(thread_type == THREAD_TYPE_PM && session.Path[1] != URL_PM) {

		pageUnableToOpen(session, "thread")
		return ""
	}

	if thread_type == THREAD_TYPE_ARTICLE && article_id != nil && article_id.(int64) > 0 {
		// if the thread is a comments section for an article,
		// then we'll make it a special section titles, descriptions and
		// breadcrumbs all having something meaningful.
		article_preview = "1"
		session.Article.ID = uint(article_id.(int64))
		session.Article.Title.Value = string(article_title.([]uint8))
		session.Article.Description.Value = trim(rx_purge_tags.ReplaceAllString(string(article_desc.([]uint8)), ""))
		link_url = session.Article.GetURL("article")

		session.Topic.ID = uint(topic_id.(int64))
		session.Topic.Title.Value = string(topic_title.([]uint8))
		session.Topic.Description.Value = string(topic_desc.([]uint8))

		session.Page.Section = &session.Special
		trimPString(&session.Article.Description.Value, uint(CORE_THREAD_LINK_DESC_MAX_C))

		session.Thread.Title.Value = "Comments"
		session.Thread.Description.Value = "Comments: " + session.Article.Title.Value
		session.Special.Title.Value = session.Thread.Description.Value
		session.Special.Description.Value = session.Article.Description.Value
		session.Special.ID = session.Thread.ID

	} else {
		if link_url != "" {
			article_preview = "1"
			link_desc = strings.Replace(link_desc, "<", `&lt;`, -1)
			session.Article.Description.Value = strings.Replace(link_desc, ">", `&gt;`, -1)
		}

		if forum_id != nil {
			session.Forum.ID = uint(forum_id.(int64))
			session.Forum.Title.Value = string(forum_title.([]uint8))
			session.Forum.Description.Value = string(forum_desc.([]uint8))
		} else {
			session.Forum.Title.Value = "root"
		}
	}

	if thread_type == THREAD_TYPE_PM {
		// TODO: do something
	}

	// TODO: description and titles for other types of threads
	/*	session.Thread.ID = uint(forum_id.(int64))
		session.Thread.Description.Value = session.Thread.Title.Value
		session.Page.Section = &session.Special
		session.Special.ID = session.Thread.ID
		session.Special.Title.Value = string(forum_title.([]uint8))
		session.Special.Description.Value = string(forum_desc.([]uint8))
	*/

	//if last_visit != nil {
	//	t_last_visit = dateParse(string(last_visit.([]uint8)), "pageThread")
	//	debugLog(string(last_visit.([]uint8)))
	//}

	if session.RealTime() {
		session.PostProcInc += "<script>webSockInit('/thread/" + Itoa(session.Thread.ID) + "', 'real-time-comments');</script>"
	}

	session.ReadComments = session.ReadComments.Make(s_rc)

	// post a comment?
	success, f, redirect := postComment(session)
	if redirect != "" {
		return redirect
	}
	if success {
		if thread_model == THREAD_MODEL_FLAT {
			return fmt.Sprintf("%s/page%d#comment%d", session.r.URL, session.Page.nPageCount, f.CommentID)
		} else {
			return fmt.Sprintf("%s?id=%d#comment%d", session.r.URL, f.ParentID, f.CommentID)
		}
	}

	session.Page.Content += session.layout.ThreadTemplateBC

	cnew, _ := Atoui(session.GetCookie("cnew").Value)
	forumViewThread(session, thread_model, parent_id, cnew)
	forumReplyForm(session, thread_locked, f, thread_type)
	if cnew != 0 {
		session.SetCookie("cnew", "", -1)
	}

	session.Page.Content += session.layout.ThreadTemplateAC

	session.Page.Content = strings.Replace(session.Page.Content, "<_article|preview>", article_preview, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_article|id>", Itoa(session.Article.ID), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_article|title>", session.Article.Title.HTMLEscaped(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_article|url|full>", link_url, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_article|url|short>", trimString(link_url, CORE_SOURCE_URL_CROP).HTMLEscaped(), -1)
	desc := strings.Replace(session.Article.Description.HTMLEscaped(), "\n", "<br/>", -1)
	desc = strings.Replace(desc, "  ", "&nbsp;", -1)
	desc = strings.Replace(desc, "\t", "&nbsp;&nbsp;&nbsp;&nbsp;", -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_article|description>", desc, -1)

	session.Page.Content += fmt.Sprintf(`<script>var page=%d,max_pages,tid=%d,t='%s',tname='%s';window.onscroll=function() {detectBottom();};</script>`,
		session.Page.nPageCurrent, session.Thread.ID, session.Path[1], session.Thread.Title.URLify())

	session.ReadComments.Export(session)

	return ""
}

func pageComment(session *Session) {
	var (
		r             Comment
		f             FormComment
		err           error
		thread_locked string
		thread_model  string
		thread_type   string
	)
	session.Page.Section = &session.Special

	r.ID, _ = Atoui(session.Path[2])
	err = dbSelectRow(session, true, dbQueryRow(session, SQL_SHOW_SINGLE_COMMENT,
		r.ID, session.User.Permissions.RegEx(), session.User.Permissions.RegEx()),
		&r.ParentID, &r.User.ID, &r.User.Name.Alias.Value, &r.User.Name.First.Value, &r.User.Name.Full.Value, &r.User.AvatarValue,
		&r.Content, &r.Cached, &r.Created, &r.Updated, &r.Karma, &r.ThreadID, &thread_locked, &session.Thread.Title.Value, &thread_model, &thread_type)

	// pretty out of db sessions error message
	if dbOutOfSessions(session, err) {
		return
	}

	if err != nil {
		// 404 if no comment found
		session.Forum.Title.Value = "Comment not found"
		page404(session, "comment")
		return
	}

	session.Special.Title.Value = "@" + r.User.Name.Short().HTMLEscaped()
	session.Special.Description.Value = trimString(r.Content, 600).HTMLEscaped()
	session.Thread.ID = r.ThreadID
	r.ThreadModel = &thread_model
	session.Page.Content = strings.Replace(session.layout.ShowCommentTemplate, "<_comment|id>", Itoa(r.ID), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_comment>", *forumComment(session, &r, THREAD_MODEL_FLAT, 0), 1)

	f.ParentID = r.ID
	forumReplyForm(session, thread_locked, &f, thread_type)

	loadHeaderFooter(session)
	session.PostProcInc += `<script>commentHideReply();</script></body>`
}

func pageUser(session *Session) {
	var (
		u      User
		joined time.Time
		p      UserPreferences
	)
	u.ID, _ = Atoui(session.Path[2])

	session.Page.Section = &session.Special

	// load user
	err := dbSelectRow(session, true, dbQueryRow(session, SQL_SHOW_USER, u.ID, u.ID),
		&u.Name.Alias.Value, &u.Name.First.Value, &u.Name.Full.Value, &u.Description.Value,
		&u.Email.Value, &u.JoinDate, &u.Enabled, &u.Karma,
		&u.Twitter.Name, &u.GooglePlus.ID, &u.Facebook.ID,
		&u.AvatarValue, &u.Permissions.str,
		&p.PublicEmail, &p.PublicTwitter, &p.PublicGooglePlus, &p.PublicFacebook)

	if err != nil {
		// pretty out of db sessions error message
		if dbOutOfSessions(session, err) {
			return
		}

		// 404 if no article found
		page404(session, "user")
		return
	}

	p.ExportPreferences(&u)

	session.Special.Title.Value = "@" + u.Name.Short().HTMLEscaped()
	session.Special.Description.Value = trimString(u.Description.Value, 600).HTMLEscaped()

	cmsBBDecode(&u.Description.Value, session)
	joined = dateParse(u.JoinDate, "pageUser")

	session.Page.Content = strings.Replace(session.layout.ShowUserTemplate, "<_user|id>", Itoa(u.ID), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|name|short>", u.Name.Short().HTMLEscaped(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|name|long>", u.Name.Long().HTMLEscaped(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|description>", u.Description.Value, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|avatar|small>", u.AvatarSmall(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|avatar|large>", u.AvatarLarge(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|karma>", Itoa(u.Karma), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|joined>", joined.Format(DATE_FMT_THREAD), -1)

	session.Page.Content = strings.Replace(session.Page.Content, "<_user|email>", u.Email.HTMLEscaped(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|twitter>", u.Twitter.Name, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|facebook>", u.Facebook.ID, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|google plus>", u.GooglePlus.ID, -1)
}

func pageMe(session *Session) {
	var (
		u      User
		joined time.Time
		p      UserPreferences
	)
	u.ID = session.User.ID

	session.Page.Section = &session.Special

	// load user
	err := dbSelectRow(session, true, dbQueryRow(session, SQL_SHOW_USER, u.ID, u.ID),
		&u.Name.Alias.Value, &u.Name.First.Value, &u.Name.Full.Value, &u.Description.Value,
		&u.Email.Value, &u.JoinDate, &u.Enabled, &u.Karma,
		&u.Twitter.Name, &u.GooglePlus.ID, &u.Facebook.ID,
		&u.AvatarValue, &u.Permissions.str,
		&p.PublicEmail, &p.PublicTwitter, &p.PublicGooglePlus, &p.PublicFacebook)

	if err != nil {
		// pretty out of db sessions error message
		if dbOutOfSessions(session, err) {
			return
		}

		// 404 if no article found
		page404(session, "user")
		return
	}

	session.Special.Title.Value = "@" + u.Name.Short().HTMLEscaped()
	session.Special.Description.Value = trimString(u.Description.Value, 600).HTMLEscaped()

	cmsBBDecode(&u.Description.Value, session)
	joined = dateParse(u.JoinDate, "pageUser")

	session.Page.Content = strings.Replace(session.layout.ShowMeTemplate, "<_user|id>", Itoa(u.ID), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|name|short>", u.Name.Short().HTMLEscaped(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|name|long>", u.Name.Long().HTMLEscaped(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|description>", u.Description.Value, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|avatar|small>", u.AvatarSmall(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|avatar|large>", u.AvatarLarge(), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|karma>", Itoa(u.Karma), -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|joined>", joined.Format(DATE_FMT_THREAD), -1)

	session.Page.Content = strings.Replace(session.Page.Content, "<_user|twitter>", u.Twitter.Name, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|facebook>", u.Facebook.ID, -1)
	session.Page.Content = strings.Replace(session.Page.Content, "<_user|google plus>", u.GooglePlus.ID, -1)
}

func pageLogin(session *Session) string {
	// lets not let people login if they're already logged in
	if session.ID != "" {
		return SITE_HOME_URL
	}

	var (
		success  bool
		f        *FormLoginRegister
		required map[bool]string
	)

	f = new(FormLoginRegister)
	required = make(map[bool]string)
	required[false] = "optional"
	required[true] = "required"

	session.Page.Section = &session.Special
	session.Special.Title.Value = "Login / Register"
	session.Special.Description.Value = SITE_DESCRIPTION

	if session.Path[2] == "post" {
		session.Page.Content += `
        <p class="auth_err">You need to login first. But don't worry, your comment has been saved in your browser's local files.</p>`
	}

	// social media logins
	if ENABLE_FACEBOOK_LOGIN && session.Path[2] == "facebook" {
		return facebookLogin(session)
	}

	// login
	if session.GetPost("login").Value == YES {
		success, f = postLogin(session)
		if success {
			return getReferrer(session)
		}
	}

	session.Page.Content += fmt.Sprintf(`<p id="login-message" class="auth_err">%s</p>
	    <h1>Login</h1>
	    Already a member? Login here:
	    <form action="<_url|home>login#login-message" method="post" accept-charset="UTF-8">

	        <table class="register">
	            <tr>
	                <td class="reg_required">Username</td>
	                <td></td>
	                <td><input type="text" name="username" class="textbox" value="%s" maxlength="%d"/></td>
	                <td class="reg_notice">%s</td>
	            </tr>
	            <tr>
	                <td class="reg_required">Password</td>
	                <td></td>
	                <td><input type="password" name="password" class="textbox" maxlength="%d"/></td>
	                <td class="reg_notice">%s</td>
	            </tr>
	            <tr>
	                <td class="reg_required"><button name="login" value="%s" type="submit" class="btn">Login!</button></td>
	                <td></td>
	                <td colspan="2"><a href="javascript:alert('Not yet implemented');">(Request password reset)</a></td>
	            </tr>
	        </table>
	    </form>`, f.ErrorMessage, f.Username.Value, CORE_USERNAME_MAX_CHARS, f.Username.Error, CORE_PASSWORD_MAX_CHARS, f.Password.Error, YES)

	// facebook login
	if ENABLE_FACEBOOK_LOGIN {
		session.Page.Content += `<div>
			<h3>or you can sign in with one of the following social media accounts:</h3>
			<div id="fb_login">
				<a href="<_url|home>login/facebook/" title="Log in to <_site|name> via facebook">
					<img src="<_cdn|images>layout/facebook.png" alt="Log in to <_site|name> via facebook">
				</a>
			</div></div>
			<div class="clear"></div>`
	}

	// register
	f = new(FormLoginRegister)
	session.Page.Content += `<p></p><hr/>`
	if session.GetPost("register").Value == YES {
		success, f = postRegister(session)
		if success {
			return getReferrer(session)
		}
	}

	session.Page.Content += fmt.Sprintf(`<p id="register-message" class="auth_err">%s</p>
	    <h1>Register</h1>
	    New members register here:
	    <form action="<_url|home>login#register-message" method="post" autocomplete="off" accept-charset="UTF-8">

	        <table class="register">
	            <tr>
	                <td class="reg_required">Username</td>
	                <td></td>
	                <td><input type="text" name="username" class="textbox" value="%s"
                        onkeyup="fieldHelper(this, 'r_username', '%s');"
                        onfocus="fieldHelper(this, 'r_username', '%s');"
                        maxlength="%d"/></td>
	                <td><span class="reg_notice" id="r_username">%s</span></td>
	            </tr>`,
		f.ErrorMessage, f.Username.Value,
		lang(session, "required_field").JSEscaped(), lang(session, "required_field").JSEscaped(),
		CORE_USERNAME_MAX_CHARS, f.Username.Error)

	session.Page.Content += fmt.Sprintf(`
	            <tr>
	                <td class="reg_required">Password</td>
	                <td></td>
	                <td><input type="password" name="password1" class="textbox"
                        onkeyup="fieldHelper(this, 'r_password1', '%s');"
                        onfocus="fieldHelper(this, 'r_password1', '%s');"
                        maxlength="%d"/></td>
	                <td><span class="reg_notice" id="r_password1">%s</span></td>
	            </tr>
	            <tr>
	                <td class="reg_required">Password (re-enter)</td>
	                <td></td>
	                <td><input type="password" name="password2" class="textbox"
                        onkeyup="fieldHelper(this, 'r_password2', '%s');"
                        onfocus="fieldHelper(this, 'r_password2', '%s');"
                        maxlength="%d"/></td>
	                <td><span class="reg_notice" id="r_password2">%s</span></td>
	            </tr>`,
		lang(session, "required_field").JSEscaped(), lang(session, "required_field").JSEscaped(), CORE_PASSWORD_MAX_CHARS, f.Password.Error,
		lang(session, "required_field").JSEscaped(), lang(session, "required_field").JSEscaped(), CORE_PASSWORD_MAX_CHARS, f.Password.Error)

	session.Page.Content += fmt.Sprintf(`
	            <tr>
	                <td class="reg_%s">Email Address</td>
	                <td></td>
	                <td><input type="email" name="email" class="textbox" value="%s"
                        onkeyup="fieldHelper(this, 'r_email', '%s');"
                        onfocus="fieldHelper(this, 'r_email', '%s');"
                        maxlength="%d"/></td>
	                <td><span class="reg_notice" id="r_email">%s</span></td>
	            </tr>`,
		required[CORE_REG_EMAIL_REQUIRED], f.Email.Value,
		lang(session, required[CORE_REG_EMAIL_REQUIRED]+"_field").JSEscaped(), lang(session, required[CORE_REG_EMAIL_REQUIRED]+"_field").JSEscaped(),
		CORE_EMAIL_MAX_CHARS, f.Email.Error)

	session.Page.Content += fmt.Sprintf(`
                <tr>
	                <td class="reg_%s">Twitter Username</td>
	                <td>@</td>
	                <td><input type="text" name="twitter" class="textbox" value="%s"
                        onkeyup="fieldHelper(this, 'r_twitter', '%s');"
                        onfocus="fieldHelper(this, 'r_twitter', '%s');"
                        maxlength="%d"/></td>
	                <td><span class="reg_notice" id="r_twitter">%s</span></td>
	            </tr>`,
		required[CORE_REG_TWITTER_REQUIRED], f.Twitter.Value,
		lang(session, required[CORE_REG_TWITTER_REQUIRED]+"_field").JSEscaped(), lang(session, required[CORE_REG_TWITTER_REQUIRED]+"_field").JSEscaped(),
		CORE_TWITTER_MAX_CHARS, f.Twitter.Error)

	session.Page.Content += fmt.Sprintf(`
	            <tr>
	                <td class="reg_required"><button name="register" value="%s" type="submit" class="btn">Register!</button></td>
	                <td></td>
	                <td colspan="2">
                        <a href="javascript:showHiddenContainer('privacy_copyright');"
                            title="View privacy and copyright notes.">(Privacy and copyright notes)</a>
                    </td>
	            </tr>
	        </table>
	    </form>`, YES)

	session.Page.Content += `
	    <div id="privacy_copyright" style="display: none;">
	        <p></p><hr/>
	        <h1>About your data</h1>
	        <_layout|privacy>
		</div>`

	return ""
}

func pageLogout(session *Session) string {
	if matchTokens(session, session.r.URL.Query().Get("t")) {
		session.WriteSessionCookies(false)
	}

	return getReferrer(session)
}

func pageSwitchPlatform(session *Session) string {
	if session.Path[2] == "mobile" || session.Path[2] == "desktop" {
		session.SetCookie("platform", session.Path[2], time.Duration(CORE_PLATFORM_PREF_AGE))
	}
	return getLast(session)
}

func pageRedirect(session *Session, url string) {
	setStatus(session, 302)
	session.w.Header().Set("Location", url)
	session.w.Header().Set("Cache-Control", "no-cache")
	session.Page.Content = fmt.Sprintf(`
	<html>
		<head>
			<meta http-equiv="cache-control" content="no-cache" />
			<meta http-equiv="Refresh" content="1; url=%s" />
		</head>
		<body>
			<script type="text/javascript" language="javascript">window.location='%s';</script>
			<h1>This page has been redirected:</h1>
			<p>If you're not automatically redirected, then please click to the following link to continue <a href="%s">%s</a></p>
		</body>
	</html>%s`, url, url, url, url, "\n")
	//session.w.Write([]byte(session.Page.Content))
	writeBody(session)
}

func pageUnableToOpen(session *Session, section string) {
	setStatus(session, http.StatusNotFound)

	if session.Page.Title().Value == "" {
		session.Page.Section.Title.Value = "Cannot open " + section
	}
	session.Page.Section.Description.Value = SITE_DESCRIPTION
	if session.Page.Content == "" {
		session.Page.Content = fmt.Sprintf(`<h1>Cannot open %s!</h1><p>Either this %s does not exist or you do not have permission to access it.</p>`, section, section)
	} else {
		session.Page.Content = `<h1>Access denied!</h1><p>` + session.Page.Content + `</p>`
	}
}

func pageDenied(session *Session, section string) {
	setStatus(session, http.StatusForbidden)

	if session.Page.Title().Value == "" {
		session.Page.Section.Title.Value = "Permission denied"
	}
	session.Page.Section.Description.Value = SITE_DESCRIPTION
	if session.Page.Content == "" {
		session.Page.Content = fmt.Sprintf(`<h1>Permission denied!</h1><p>You do not have permission to access this %s.</p>`, section)
	} else {
		session.Page.Content = `<h1>Permission denied!</h1><p>` + session.Page.Content + `</p>`
	}
}

func page404(session *Session, section string) {
	setStatus(session, http.StatusNotFound)

	if session.Page.Title().Value == "" {
		session.Page.Section.Title.Value = "Resource not found"
	}
	session.Page.Section.Description.Value = SITE_DESCRIPTION
	if session.Page.Content == "" {
		session.Page.Content = fmt.Sprintf(`<h1>Something went wrong 😞</h1><p>The %s you've requested does not exist.</p>`, section)
	} else {
		session.Page.Content = `<h1>Something went wrong 😞</h1><p>` + session.Page.Content + `</p>`
	}
}

func dbOutOfSessions(session *Session, err error) bool {
	if err != nil {
		if err.Error() == "Error 1040: Too many connections" {
			pageTooManyConnections(session)
			return true
		}
	}
	return false
}

func pageTooManyConnections(session *Session) {
	session.Page.Section.Title.Value = "Site experiencing heavy traffic"
	session.Page.Section.Description.Value = "Site experiencing heavy traffic"
	session.Page.Content = fmt.Sprintf(`<h1>Database overloaded 😞</h1>
		<p>%s is experiencing unusually high traffic.</p>
		<p>Please hit <a href="%s" title="%s">refresh</a> and try again in a few moments.</p>`,
		SITE_NAME, session.r.URL.RequestURI(), SITE_NAME+": "+SITE_DESCRIPTION)
}

func pageRobots(session *Session) {
	session.w.Header().Set("Content-Type", "text/plain")
	if CORE_ALLOW_ROBOTS {
		session.Page.Content = "User-agent: *\nAllow: /\n"
	} else {
		session.Page.Content = "User-agent: *\nDisallow: /\n"
	}
	session.ResponseSize = len(session.Page.Content)
	//session.w.Write([]byte(session.Page.Content))
	writeHeaders(session)
	//session.w.Header().Set("Content-Type", "text/plain")
	writeBody(session)
}

func pagePing(session *Session) {
	session.w.Header().Set("Content-Type", "text/plain")

	session.Page.Content = "PONG\n"

	session.ResponseSize = len(session.Page.Content)
	//session.w.Write([]byte(session.Page.Content))
	writeHeaders(session)
	//session.w.Header().Set("Content-Type", "text/plain")
	writeBody(session)
}

func pageUserList(session *Session) {
	session.Page.Section = &session.Special

	session.Special.Title.Value = "user list"
	session.Special.Description.Value = "user list"

	session.Page.Content = session.layout.SelectUserTemplate
}
