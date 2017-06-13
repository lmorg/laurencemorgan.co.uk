// level10fireball
package main

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CMS_NAME      string = "Level 10 Fireball"
	CMS_URL       string = "http://h4ck.in"
	CMS_VERSION   string = "1.3.5200 ALPHA"
	CMS_COPYRIGHT string = "© Laurence 2012-2016"
)

////////////////////////////////////////////////////////////////////////////////

type Session struct {
	ID    string
	Token string

	User            User
	Mobile          bool
	Theme           string
	Language        string
	Page            PageSession // current page being rendered
	Special         Section
	Article         Section
	Topic           Section
	Thread          Section
	Forum           Section
	ReadComments    ReadComments
	UnreadComments  map[uint]bool
	GalleryID       uint     // just in case there's multiple galleries per page
	Path            []string // the URL is broken down into an array for ease
	File            string   // last item in path
	PostProcInc     string   // misc styling included in page footer (usually geenrated CSS and Javascript)
	Variables       map[string]string
	Now             time.Time
	w               http.ResponseWriter
	r               *http.Request
	db              *sql.DB
	layout          *Layout
	Status          int
	ResponseSize    int
	RealTimeRequest bool
}

// TODO: does this even work. what if createtoken == 99.99, date ticks over in seconds.
// Also, why am I returning the token /AND/ storing it in session?
func (session Session) CreateToken(modifier int) (token string) {
	hours := int(time.Duration(time.Now().UnixNano()).Hours()/CORE_TOKEN_AGE) + modifier
	token = validationHash(&session, strconv.Itoa(hours))
	//debugLog(fmt.Sprintf("Creating token for %d, %s", hours, token))
	session.Token = token
	return
}

func (session Session) CreateSessionID() string {
	return validationHash(&session, session.User.Name.Long().Value+time.Now().Format(DATE_FMT_ARTICLE))
	// TODO: should be session.Now.Format()
}

func (session Session) WriteSessionCookies(login bool) {
	if login {
		session.SetCookie("session", session.ID, time.Duration(CORE_SESSION_AGE))
		session.SetCookie("user", session.User.Hash, time.Duration(CORE_SESSION_AGE))
	} else {
		session.SetCookie("session", "", -1)
		session.SetCookie("user", "", -1)
	}
}

func (session Session) SetCookie(key, value string, max_age time.Duration) {
	if session.RealTimeRequest { // just in case real time requests are trying to write session cookies.
		return
	}

	var (
		cookie  http.Cookie
		expires time.Time
	)

	//expires = time.Now().Add(max_age * time.Second)
	expires = session.Now.Add(max_age * time.Second)

	cookie.Domain = COOKIE_DOMAIN
	cookie.Expires = expires
	cookie.HttpOnly = true
	cookie.MaxAge = int(max_age)
	cookie.Name = url.QueryEscape(key)
	cookie.Path = COOKIE_HOME_PATH
	cookie.Secure = false
	cookie.Value = url.QueryEscape(value)

	http.SetCookie(session.w, &cookie)
}

func (session Session) GetCookie(key string) DisplayText {
	s, _ := url.QueryUnescape(key)
	cookie, err := session.r.Cookie(s)
	if err != nil {
		//isErr(&session, err, true, "reading cookie", "GetCookie")
		return (DisplayText{""})
	}
	s, _ = url.QueryUnescape(cookie.Value)
	return (DisplayText{s})
}

// I can't believe I have to write this fucking kludge
func (session Session) GetPost(key string) DisplayText {
	if session.r.FormValue(key) == session.r.URL.Query().Get(key) {
		return (DisplayText{""})
	}
	return (DisplayText{session.r.FormValue(key)})
}

func (session Session) GetQueryString(key string) DisplayText {
	return (DisplayText{session.r.URL.Query().Get(key)})
}

func (session Session) RealTime() bool {
	if (!session.Mobile && ENABLE_REAL_TIME_DESKTOP) ||
		(session.Mobile && ENABLE_REAL_TIME_MOBILE) {
		return true
	}
	return false
}

func NewSession(w http.ResponseWriter, r *http.Request, layout *Layout) (session Session) {
	session.Now = time.Now() // first job so we can capture proc times

	session.Page.Images = append(session.Page.Images, "http:"+SITE_SOCIAL_MEDIA_IMAGE)
	session.Variables = make(map[string]string)
	session.UnreadComments = make(map[uint]bool)

	session.w = w
	session.r = r
	session.layout = layout
	session.Status = 200

	session.Theme = "desktop"
	session.Language = "en"
	// check if moble. TODO: at some point I'm going to turn mobile and desktop into themes
	if session.GetCookie("platform").Value == "mobile" || session.GetCookie("platform").Value == "desktop" {
		if session.GetCookie("platform").Value == "mobile" {
			session.Mobile = true
			session.Theme = "mobile"
		}
		session.SetCookie("platform", session.GetCookie("platform").Value, time.Duration(CORE_PLATFORM_PREF_AGE))
	} else if len(session.r.UserAgent()) > 3 && (rx_mobile1.MatchString(session.r.UserAgent()) || rx_mobile2.MatchString(session.r.UserAgent()[1:4])) {
		session.Mobile = true
		session.Theme = "mobile"
	}

	return session
}

func autoLogin(session *Session) {
	session_id := session.GetCookie("session").Value
	user_hash := session.GetCookie("user").Value
	if session_id != "" && user_hash != "" {
		var (
			count       byte
			user_id     interface{} //int
			alias       interface{} //string
			first_name  interface{} //string
			full_name   interface{} //string
			join_date   interface{} //string
			permissions interface{} //string
			salt        interface{} //string
		)

		err := dbSelectRow(session, true, dbQueryRow(session, SQL_AUTO_LOGIN, session_id, user_hash),
			&count, &user_id, &alias, &first_name, &full_name, &join_date, &permissions, &salt)

		if err != nil {
			return
		}

		if count == 1 {
			session.ID = session_id
			session.User.Hash = user_hash
			session.User.ID = uint(user_id.(int64))
			session.User.Name.Alias.Value = string(alias.([]byte))
			session.User.Name.First.Value = string(first_name.([]byte))
			session.User.Name.Full.Value = string(full_name.([]byte))
			session.User.JoinDate = string(join_date.([]byte))
			session.User.Permissions.str = string(permissions.([]byte))
			session.User.Salt = string(salt.([]byte))
			session.WriteSessionCookies(true)
			session.Token = session.CreateToken(0)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type Version struct {
	Major    byte
	Minor    byte
	Revision string
	Version  string
}

func (v Version) ResString() string {
	if !DEBUG {
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	} else {
		return fmt.Sprintf("%d.%d.%s", v.Major, v.Minor, strings.Replace(v.Revision, " ", "", -1))
	}
}

func NewVersion() (version Version) {
	s := strings.Split(CMS_VERSION, ".")
	version.Major, _ = Atob(s[0])
	version.Minor, _ = Atob(s[1])
	version.Revision = s[2]
	version.Version = fmt.Sprintf("%d.%d (build %s)", version.Major, version.Minor, version.Revision)
	return
}

////////////////////////////////////////////////////////////////////////////////

var (
	version               Version
	rx_mobile1            *regexp.Regexp
	rx_mobile2            *regexp.Regexp
	rx_articles_template  *regexp.Regexp
	rx_topics_template    *regexp.Regexp
	rx_forums_template    *regexp.Regexp
	rx_subforums_template *regexp.Regexp
	rx_threads_template   *regexp.Regexp
	rx_gthumbs_template   *regexp.Regexp
)

func init() {
	version = NewVersion()

	// lets compile the regex up front for additional performance
	rx_mobile1 = regexCompile(`(?i)(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|iris|kindle|lge |maemo|midp|mmp|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows (ce|phone)|xda|xiino`)
	rx_mobile2 = regexCompile(`(?i)1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-`)

	rx_articles_template = regexCompile(`(?s)<_articles\|(begin|1)>(.*?)<_articles\|(end|0|)>`)
	rx_topics_template = regexCompile(`(?s)<_topics\|(begin|1)>(.*?)<_topics\|(end|0|)>`)
	rx_forums_template = regexCompile(`(?s)<_forums\|(begin|1)>(.*?)<_forums\|(end|0|)>`)
	rx_subforums_template = regexCompile(`(?s)<_subforums\|(begin|1)>(.*?)<_subforums\|(end|0|)>`)
	rx_threads_template = regexCompile(`(?s)<_threads\|(begin|1)>(.*?)<_threads\|(end|0|)>`)
	rx_gthumbs_template = regexCompile(`(?s)<_thumbs\|(begin|1)>(.*?)<_thumbs\|(end|0|)>`)
}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	var (
		session Session
		//access_log AccessLog
	)

	defer func() {
		dbClose(session.db) // if db connection isn't already closed...
		if session.Path[1] != "PING" {
			accessLog(&session) // log http requests
		}
	}()

	// create session
	session = NewSession(w, r, &live_layout)

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
	} else {

		autoLogin(&session)

		// now lets start serving content based upon the URL
		var redirect string

		switch session.Path[1] {
		case "", "g":
			pageHome(&session)

		case "a", "article":
			pageArticle(&session)

		case "l", "list":
			pageListArticles(&session)

		case "topics": //section / subject?
			pageAllTopics(&session)

		case "c", "comment":
			pageComment(&session)

		case "pm":
			redirect = pagePMs(&session)

		case "t", "thread":
			redirect = pageThread(&session)

		case "f", "forum", "forums":
			redirect = pageAllForums(&session)

		case "u", "user": // member?
			pageUser(&session)

		case "me":
			pageMe(&session)

		case "ul":
			pageUserList(&session)

		case "login":
			redirect = pageLogin(&session)

		case "logout":
			redirect = pageLogout(&session)

		case "platform":
			redirect = pageSwitchPlatform(&session)

		case "root":
			redirect = pageRoot(&session)

		//case "favicon.ico":
		//	redirect = URL_IMAGE_PATH + "layout/favicon.ico"

		case "ajax":
			pageAJAX(&session)
			return

		case "robots.txt":
			pageRobots(&session)
			return

		case "ping", "PING":
			pagePing(&session)
			return

		case "templates", "conf", "bin":
			session.Page.Section = &session.Special
			page404(&session, "page")

		default:
			if CORE_WEB_SERVER_MODE && len(session.File) > 0 && session.File[:1] != "." {
				dbClose(session.db) // we don't need a database to read / write static files
				pageWebServerMode(&session)
				if session.Status == 200 {
					return
				}
			} else {
				session.Page.Section = &session.Special
				page404(&session, "page")
			}
		}

		// close database (no longer needed)
		dbClose(session.db)

		// page requested a redirect
		if redirect != "" {
			pageRedirect(&session, redirect)
			return
		}
	}

	session.PostProcInc += fmt.Sprintf(`<script>var token='%s';</script>`, session.Token)
	if session.User.ID == 0 {
		session.PostProcInc += `<style>.guest_vis{visibility:hidden;}.guest_dis{display:none;}</style>`
	}

	// create HTTP headers
	writeHeaders(&session)
	session.SetCookie("last", SITE_HOME_URL+r.URL.Path[1:], time.Duration(CORE_REFERRER_AGE))
	if CORE_REFERRER_COOKIE && len(session.Path[1]) >= 3 && session.Path[1][:3] != "log" {
		session.SetCookie("ref", SITE_HOME_URL+r.URL.Path[1:], time.Duration(CORE_REFERRER_AGE))
	}

	// load applicable page headers and footers
	loadHeaderFooter(&session)

	// parse the page
	cmsLayout(&session.Page.Header, &session)

	if session.Page.Section != &session.Article {
		cmsLayout(&session.Page.Content, &session) // we only layout parse the content here it's not an article
	} //                                           // (essentially stopping content editors from using layout tags)
	cmsLayout(&session.Page.Footer, &session)

	cms(&session.Page.Content, &session) // content first as some header bits are derived
	cms(&session.Page.Header, &session)
	cms(&session.Page.Footer, &session)

	// minify code
	if (!session.Mobile && CORE_MINIFY_DESKTOP) || (session.Mobile && CORE_MINIFY_MOBILE) {
		cmsMinify(&session.Page.Header, &session)
		cmsMinify(&session.Page.Content, &session)
		cmsMinify(&session.Page.Footer, &session)
	}

	// finally transmit this dynamic page
	writeBody(&session)
}

func setStatus(session *Session, status int) {
	session.w.WriteHeader(status)
	session.Status = status
}

func writeBody(session *Session) {
	session.ResponseSize = len(session.Page.Header) + len(session.Page.Content) + len(session.Page.Footer)
	//session.ResponseSize = 1337
	if session.Status == 200 && strings.Contains(session.r.Header.Get("Accept-Encoding"), "gzip") {
		gz := gzip.NewWriter(session.w)
		_, err := gz.Write([]byte(session.Page.Header + session.Page.Content + session.Page.Footer))
		isErr(session, err, true, "Creating gzip writer", "pageHandler")
		isErr(session, gz.Close(), true, "Closing gzip writer", "pageHandler")
	} else {
		_, err := session.w.Write([]byte(session.Page.Header + session.Page.Content + session.Page.Footer))
		isErr(session, err, true, "HTTP writer", "pageHandler")
	}
}

func writeHeaders(session *Session) {
	session.w.Header().Set("CachCachCacheeCachee-Control", fmt.Sprintf("max-age=%d", CORE_CACHE_CONTROL))
	session.w.Header().Set("Content-Language", session.Language)
	session.w.Header().Set("Server", "Level 10 Fireball")
	session.w.Header().Set("X-Powered-By", "Late nights and lot's of Scotch")
	session.w.Header().Set("X-Frame-Options", "sameorigin")
	session.w.Header().Set("X-XSS-Protection", "1; mode=block")
	session.w.Header().Set("X-Content-Type-Options", "nosniff")
	if strings.Contains(session.r.Header.Get("Accept-Encoding"), "gzip") && session.Status == 200 {
		session.w.Header().Set("Content-Encoding", "gzip")
	}
	//session.w.Header().Set("Connection", "close")
	//session.w.Header().Set("Vary", "Accept-Encoding")
	//session.w.Header().Set("Transfer-Encoding", "chunked")
	session.w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func loadHeaderFooter(session *Session) {
	if !session.Mobile {
		session.Page.Header = session.layout.Desktop.Header // + "<p>"
		session.Page.Footer = session.layout.Desktop.Footer
	} else {
		session.Page.Header = session.layout.Mobile.Header // + "<p>"
		session.Page.Footer = session.layout.Mobile.Footer
	}
}

func loadEnvironmentFromDB(layout *Layout) {
	var (
		session          Session // = NewSession(nil, nil, nil)
		errp             error
		label, desc, url string
	)
	layout.GalleryTemplates = make(map[string]GalleryTemplate)

	_, session.db = dbOpen()

	// load MIME types if using level 10 as a web server
	confFileTypes()

	// preload quotes (TODO: rename to fortune)
	var fortunes string
	readTemplate(&session, "fortune", &fortunes)
	fortunes = strings.Replace(html.EscapeString(fortunes), "|", "<br/>", -1)
	layout.Quotes = strings.Split(fortunes, "\n")

	// preload embedded content
	readTemplate(&session, "privacy", &layout.PrivacyTemplate)
	readTemplate(&session, "bbcode supported", &layout.BBCode)
	readTemplate(&session, "social buttons", &layout.SocialButtons)

	// preload menubar items
	layout.Menubar = nil
	rows := dbQueryRows(&session, true, &errp, SQL_MENUBAR)
	for dbSelectRows(&session, true, &errp, rows, &label, &desc, &url) {
		layout.Menubar = append(layout.Menubar, Menubar{label, desc, url})
	}

	layout.MenubarLeft = "&lt;&nbsp;"
	layout.MenubarRight = "&nbsp;&gt;"

	// TODO: the following will eventually be pushed into the database as well (or config file (or both))

	layout.BreadcrumbSeparator = `&gt;&gt;`
	layout.nListColumns = 2
	layout.nArticlesPerTopic = 3
	layout.nItemsPerPage = 8
	layout.nCommentsPerHighlight = 3
	layout.nCommentNestsPerPage = 10
	//layout.nCommentsThreadedPerPage = 15
	layout.nCommentsFlatPerPage = 20
	layout.nCharsMobileBreadcrumbs = 18
	layout.ArticleAppendComments = COMMENTS_HIGHLIGHTS //TODO: I need to test COMMENTS_ALL works
	layout.xEmbeddedFrame = 800                        //640
	layout.yEmbeddedFrame = 600                        //390
	layout.nForumRowsColours = 2

	//layout.Desktop.RealTime = ENABLE_REAL_TIME_DESKTOP
	//layout.Mobile.RealTime = ENABLE_REAL_TIME_MOBILE

	readTemplate(&session, "header desktop", &layout.Desktop.Header)
	readTemplate(&session, "footer desktop", &layout.Desktop.Footer)
	readTemplate(&session, "header mobile", &layout.Mobile.Header)
	readTemplate(&session, "footer mobile", &layout.Mobile.Footer)

	readTemplate(&session, "article", &layout.ArticleTemplateBC)

	if strings.Contains(layout.ArticleTemplateBC, "<_comments>") {
		// split the page to make it better to output comments
		i := strings.Index(layout.ArticleTemplateBC, "<_comments>")
		layout.ArticleTemplateAC = layout.ArticleTemplateBC[i+11:]
		layout.ArticleTemplateBC = layout.ArticleTemplateBC[:i]
	}

	readTemplate(&session, "articles", &layout.ArticlesTemplate)

	layout_list_articles_template := rx_articles_template.FindString(layout.ArticlesTemplate)
	layout.ArticlesTemplate = strings.Replace(layout.ArticlesTemplate, layout_list_articles_template, "<_articles>", -1)
	layout_list_articles_submatch := rx_articles_template.FindStringSubmatch(layout_list_articles_template)
	layout.ArticlesTemplateArticle = layout_list_articles_submatch[2]

	readTemplate(&session, "pagination", &layout.PaginationTemplate)

	readTemplate(&session, "topics", &layout.TopicsTemplate)

	layout_topics_template := rx_topics_template.FindString(layout.TopicsTemplate)
	layout.TopicsTemplate = strings.Replace(layout.TopicsTemplate, layout_topics_template, "<_topics>", -1)
	layout_topics_submatch := rx_topics_template.FindStringSubmatch(layout_topics_template)
	layout.TopicsTemplateTopic = layout_topics_submatch[2]

	layout_articles_template := rx_articles_template.FindString(layout.TopicsTemplateTopic)
	layout.TopicsTemplateTopic = strings.Replace(layout.TopicsTemplateTopic, layout_articles_template, "<_articles>", -1)
	layout_articles_submatch := rx_articles_template.FindStringSubmatch(layout_articles_template)
	layout.TopicsTemplateArticle = layout_articles_submatch[2]

	var thread_template string
	readTemplate(&session, "thread", &thread_template)

	thread_template_split := strings.SplitN(thread_template, "<_comments>", 2)
	layout.ThreadTemplateBC += thread_template_split[0]
	layout.ThreadTemplateAC += thread_template_split[1]

	readTemplate(&session, "show comment", &layout.ShowCommentTemplate)
	readTemplate(&session, "comment", &layout.CommentTemplate)
	readTemplate(&session, "forums", &layout.ForumsTemplate)

	layout_forums_template := rx_forums_template.FindString(layout.ForumsTemplate)
	layout.ForumsTemplate = strings.Replace(layout.ForumsTemplate, layout_forums_template, "<_forums>", -1)
	layout_forums_submatch := rx_forums_template.FindStringSubmatch(layout_forums_template)
	layout.ForumsTemplateForum = layout_forums_submatch[2]

	layout_subforums_template := rx_subforums_template.FindString(layout.ForumsTemplateForum)
	layout.ForumsTemplateForum = strings.Replace(layout.ForumsTemplateForum, layout_subforums_template, "<_subforums>", -1)
	layout_subforums_submatch := rx_subforums_template.FindStringSubmatch(layout_subforums_template)
	layout.ForumsTemplateSubForum = layout_subforums_submatch[2]

	layout_threads_template := rx_threads_template.FindString(layout.ForumsTemplate)
	layout.ForumsTemplate = strings.Replace(layout.ForumsTemplate, layout_threads_template, "<_threads>", -1)
	layout_threads_submatch := rx_threads_template.FindStringSubmatch(layout_threads_template)
	layout.ForumsTemplateThread = layout_threads_submatch[2]

	readTemplate(&session, "show user", &layout.ShowUserTemplate)
	readTemplate(&session, "show me", &layout.ShowMeTemplate)
	readTemplate(&session, "select user", &layout.SelectUserTemplate)

	dbClose(session.db)
}
