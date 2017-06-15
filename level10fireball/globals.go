// config
package main

import "time"

var (
	/********************************
	        PERSONALISATIONS:
	********************************/

	// site hosting
	SITE_NAME        string = "untitled"
	SITE_DESCRIPTION string
	SITE_COPYRIGHT   string
	SITE_VERSION     string

	// social media metadata (eg used for open graph)
	SITE_TWITTER            string
	SITE_SOCIAL_MEDIA_IMAGE string
	DEFAULT_AVATAR          string

	/********************************
	        WEB SERVER CONFIG:
	********************************/

	// IP to listen on
	SITE_HOST string = "127.0.0.1" // best practice, listen on localhost, and then forward requests via Apache reverse proxy or load balancer

	// HTTP (non SSL/TLS)
	SITE_ENABLE_HTTP bool   = true
	SITE_HTTP_PORT   uint16 = 8080 // best practice, port > 1024 then reverse proxy from Apache / etc

	// HTTPS (SSL/TLS)
	SITE_ENABLE_TLS bool   = false // TODO: not working yet. probably best to run without TLS on localhost and reverse proxy to enable HTTPS
	SITE_TLS_PORT   uint16 = 8081
	SITE_TLS_KEY    string = "example.com.key"
	SITE_TLS_CERT   string = "example.com.crt"

	// address for home page
	SITE_HOME_URL string = "http://example.com/" // always append with a backslash

	// urls for non-dynamically generated content
	URL_IMAGE_PATH          string = "//example.com/images/"
	URL_STATIC_CONTENT_PATH string = "//example.com/"
	URL_WRITABLE_PATH       string = "//example.com/uploads/"
	PWD_WRITABLE_PATH       string = "/opt/example.com/uploads/"
	PWD_TEMPLATES_PATH      string = "/opt/example.com/templates/" // empty string == read from database
	PWD_WEB_CONTENT_PATH    string = "/opt/example.com/"
	CORE_WEB_SERVER_MODE    bool   = true // serve static content directly rather than proxied via Apache / nginx

	// cookie settings
	COOKIE_DOMAIN    string = "example.com"
	COOKIE_HOME_PATH string = "/" // always append with a backslash

	/********************************
	        SITE BACKEND CONFIG:
	********************************/

	// database access
	DB_PLATFORM string
	DB_DATABASE string
	DB_USERNAME string
	DB_PASSWORD string

	// DB_CON_TYPE values:
	//		tcp  == tcp/IP connection
	//		unix == unix socket connection
	DB_CON_TYPE string = "tcp"
	DB_HOSTNAME string = "127.0.0.1"
	DB_PORT_NUM uint16 = 3306
	DB_UNIX_SOC string

	// mail server settings
	SMTP_SERVER         string = "localhost"
	SMTP_PORT           int    = 25
	SMTP_SENDER_ADDRESS string = "noreply@example.com"

	/********************************
	        BONUS FEATURES:
	********************************/

	// facebook logins
	ENABLE_FACEBOOK_LOGIN bool   = false
	FACEBOOK_APP_ID       string = ""
	FACEBOOK_APP_SECRET   string = ""

	// real time updates via websockets
	ENABLE_REAL_TIME_DESKTOP bool = true
	ENABLE_REAL_TIME_MOBILE  bool = false
)

var (
	/********************************
	        CORE CONFIG:
	********************************/

	// WARNING: DO NOT EDIT UNLESS YOU KNOW WHAT YOU'RE DOING!
	// Many of these values require changes to field lengths on the database as well.
	CORE_TOKEN_AGE        float64 = 12     // number of hours per token (roughly equates to double because of the lazy cheat I'm using)
	CORE_GO_MAX_PROCS     int     = 0      // 0 == keep default assigned by Golang
	CORE_CACHE_CONTROL    int     = 0      // cache control HTTP headers - not adviceable to have above 0 since this is a dynamic website
	CORE_DB_CACHE_TIMEOUT int     = 60 * 5 // how long to cache DB results for before refresh (cache is also updated dynamically along with the database so single node systems can get away with a higher cache - albeit at cost to more RAM being consumed)
	//CORE_USER_CACHE_TIMEOUT     float64 = 60 * 15           // how long to cache user list. In seconds
	//CORE_DISK_CACHE_TIMEOUT     float64 = 60 * 60 * 24      // how long to cache disk files when running in web server mode
	CORE_DISK_CACHE_LIMIT       uint64 = 1024 * 1024 * 256 // max size of disk cache (bytes) - this is only an approximation, sizes will raise if requests are made before the webserver cache clean up executes
	CORE_DISK_CACHE_PURGER      int    = 60                // how frequently to run the cache purger (every n seconds)
	CORE_SESSION_AGE            int    = 60 * 60 * 24 * 30 // cookie age for sessions (timeout)
	CORE_PLATFORM_PREF_AGE      int    = 60 * 60 * 24 * 30 // cookie age for platform preference (timeout)
	CORE_REFERRER_COOKIE        bool   = true              // i cheat by using a referrer cookie. Can be disabled though - if it is, fallback to ref HTTP header then SITE_HOME_URL if header doesn't exist
	CORE_REFERRER_AGE           int    = 60 * 60 * 24      // how long to remember the referrer url (if enabled)
	CORE_USERNAME_MIN_CHARS     int    = 3
	CORE_USERNAME_MAX_CHARS     int    = 15
	CORE_PASSWORD_MIN_CHARS     int    = 6
	CORE_PASSWORD_MAX_CHARS     int    = 50
	CORE_EMAIL_MAX_CHARS        int    = 255
	CORE_TWITTER_MAX_CHARS      int    = 15
	CORE_SOURCE_URL_CROP        uint   = 50    //what to crop the source URL to in URL forums
	CORE_USER_AGENT_CROP        uint   = 200   // what to crop the user agent to in the DB (this will still be logged via STDOUT)
	CORE_REG_EMAIL_REQUIRED     bool   = false // email a required field for sign up?
	CORE_REG_TWITTER_REQUIRED   bool   = false // twitter a required field for sign up?
	CORE_POST_MIN_CHARS         int    = 1     // min characters per post (1 == suggested minimum)
	CORE_POST_MAX_CHARS         int    = 10000 // max characters per post
	CORE_THREAD_TITLE_MIN_C     int    = 5     // min characters for thread titles (encurages better titles)
	CORE_THREAD_TITLE_MAX_C     int    = 100   // max characters for thread titles
	CORE_THREAD_LINK_URL_MAX_C  int    = 255   // max characters for thread link URLs (link aggregator mode)
	CORE_THREAD_LINK_DESC_MIN_C int    = 5     // min characters per link description
	CORE_THREAD_LINK_DESC_MAX_C int    = 500   // max characters per link description
	CORE_THREAD_LINK_CT_MAX_C   int    = 100   // max characters per link content type
	CORE_AVATAR_SMALL_WIDTH     int    = 75
	CORE_AVATAR_SMALL_HEIGHT    int    = 75
	CORE_AVATAR_LARGE_WIDTH     int    = 300
	CORE_AVATAR_LARGE_HEIGHT    int    = 300
	CORE_MAX_COMMENT_NEST       uint   = 9     // max number of comments nested in one view (threaded)
	CORE_ENABLE_SYS_EXEC        bool   = false // enable or disable shell execute. Disabled (false) is more secure, but breaks external minification
	CORE_ALLOW_ROBOTS           bool   = false // allow or disallow robots from crawling your site
	CORE_ALLOW_REGISTRATION     bool   = true  // allow or disallow users to register
	CORE_MINIFY_DESKTOP         bool   = false // minify output HTML - desktop / mobile
	CORE_MINIFY_MOBILE          bool   = false
	CORE_MINIFY_JS_CSS          string = "" //"/usr/bin/yui-compressor $file" // appliation for minification || blank == don't minify.
)

const CORE_USER_CACHE_TIMEOUT = 60 * 15      // how long to cache user list. In seconds
const CORE_DISK_CACHE_TIMEOUT = 60 * 60 * 24 // how long to cache disk files when running in web server

var CORE_CACHE_FORUMS_ERR_TIMEOUT = 10 * time.Second
var CORE_CACHE_FORUMS_UPDATE_TIMEOUT = 60 * time.Second // eventually moved into globals

const CORE_HIDE_PM_FORUMS = true // hide PM from all forums

var (
	/********************************
	        MISC CONFIG:
	********************************/

	DATE_FMT_ARTICLE string = "Monday, 02-Jan-06 @ 15:04 (MST)"
	DATE_FMT_THREAD  string = "02-Jan-06 @ 15:04"
	//DATE_FMT_MYSQL   string = "2006-01-02T15:04:05Z07:00"
)

var PWD_CONFIG_PATH string
