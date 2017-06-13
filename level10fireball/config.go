// cache.go
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type JSONConfig map[string]interface{}

func (c JSONConfig) GetString(key string, p *string) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetString] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case string:
		*p = c[key].(string)
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetString] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetString")
	}
}

func (c JSONConfig) GetInt(key string, p *int) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetInt] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case float64:
		*p = int(c[key].(float64))
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetInt] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetInt")
	}
}

func (c JSONConfig) GetUint(key string, p *uint) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetUint] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case float64:
		*p = uint(c[key].(float64))
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetUint] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetUint")
	}
}

func (c JSONConfig) GetUint16(key string, p *uint16) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetUint16] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case float64:
		*p = uint16(c[key].(float64))
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetUint16] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetUint16")
	}
}

func (c JSONConfig) GetUint64(key string, p *uint64) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetUint64] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case float64:
		*p = uint64(c[key].(float64))
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetUint64] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetUint64")
	}
}

func (c JSONConfig) GetFloat64(key string, p *float64) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetFloat64] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case float64:
		*p = c[key].(float64)
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetFloat64] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetFloat64")
	}
}

func (c JSONConfig) GetBool(key string, p *bool) {
	if c[key] == nil {
		errLog(fmt.Sprintf(`[json config] [GetBool] "%s" is undefined]`, key))
		return
	}

	switch t := c[key].(type) {
	case bool:
		*p = c[key].(bool)
	default:
		failOnErr(errors.New(fmt.Sprintf("[json config] [GetBool] key[%s] unexpected type %T", key, t)), "(c JSONConfig) GetBool")
	}
}

//////////////////////////////////////////////////////////////////////////////////

func confReadJSON(filename, dir string) (config string) {
	rx_comment := regexCompile(`^(\s+|)//`)
	rx_include := regexCompile(`^((\s+|)#include "(.*?)")`)
	rx_math := regexCompile(`\$\{(\s+|)=(.*?)(\s+|)\}`)

	file, err := os.Open(dir + filename)
	failOnErr(err, "confReadJSON")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if rx_comment.MatchString(line) {
			continue
		}

		if rx_include.MatchString(line) {
			next_file := rx_include.FindStringSubmatch(line)[3]
			line = rx_include.ReplaceAllString(line, confReadJSON(next_file, dir))
			config += "\n" + line
			continue
		}

		match := rx_math.FindStringSubmatch(line)
		if len(match) == 4 {
			line = strings.Replace(line, match[0], strconv.Itoa(confMath(match[2])), 1)
		}

		config += "\n" + line
	}

	failOnErr(scanner.Err(), "confReadJSON")

	return
}

func confMath(s string) (result int) {
	// Experimental. Currently doesn't support negative numbers nor brackets
	// Playground: https://play.golang.org/p/BEEKCswk1I
	var (
		stack []byte
		num   int
		op    byte
	)

	math := func() {
		//fmt.Println(result, op, num)
		switch op {
		case '*':
			result = result * num
		case '/':
			result = result / num
		case '+':
			result = result + num
		case '-':
			result = result - num
		}
		num = 0
	}

	stack = append([]byte{'+'}, []byte(s)...)
	for pos, c := range stack {
		switch {
		case '0' <= c && c <= '9':
			num = (num * 10) + (int(c) - 48)

		case c == '*' || c == '/' || c == '+' || c == '-':
			math()
			op = c

		case c == ' ', c == ',':
			continue

		default:
			err := errors.New(fmt.Sprintf("Unexpected character in embedded formula: '%s', Pos:%d Char:%s.", s, pos, c))
			failOnErr(err, "confMath")
		}
	}

	math()
	return
}

func loadEnvironmentFromConf(conf_file, conf_dir string) {
	debugLog("Reading config files....")

	var config interface{}
	b := []byte(confReadJSON(conf_file, conf_dir))
	failOnErr(json.Unmarshal(b, &config), "loadEnvironmentFromConf")
	var site JSONConfig = config.(map[string]interface{})["site"].(map[string]interface{})
	var database JSONConfig = config.(map[string]interface{})["database"].(map[string]interface{})
	var daemon JSONConfig = config.(map[string]interface{})["daemon"].(map[string]interface{})
	var core JSONConfig = config.(map[string]interface{})["core"].(map[string]interface{})

	// database
	database.GetString("platform", &DB_PLATFORM)
	database.GetString("database", &DB_DATABASE)
	database.GetString("username", &DB_USERNAME)
	database.GetString("password", &DB_PASSWORD)

	database.GetString("connection type", &DB_CON_TYPE)
	database.GetString("host", &DB_HOSTNAME)
	database.GetUint16("port", &DB_PORT_NUM)
	database.GetString("unix socket", &DB_UNIX_SOC)

	site.GetString("name", &SITE_NAME)
	site.GetString("description", &SITE_DESCRIPTION)
	site.GetString("copyright", &SITE_COPYRIGHT)
	site.GetString("version", &SITE_VERSION)

	site.GetString("host", &SITE_HOST)
	site.GetBool("http enable", &SITE_ENABLE_HTTP)
	site.GetUint16("http port", &SITE_HTTP_PORT)

	site.GetBool("tls enable", &SITE_ENABLE_TLS)
	site.GetUint16("tls port", &SITE_TLS_PORT)
	site.GetString("tls key", &SITE_TLS_KEY)
	site.GetString("tls cert", &SITE_TLS_CERT)

	site.GetString("home url", &SITE_HOME_URL)
	site.GetString("image url", &URL_IMAGE_PATH)
	site.GetString("static content url", &URL_STATIC_CONTENT_PATH)
	site.GetString("writable url", &URL_WRITABLE_PATH)
	site.GetString("writable path", &PWD_WRITABLE_PATH)
	site.GetString("templates path", &PWD_TEMPLATES_PATH)
	site.GetString("web content path", &PWD_WEB_CONTENT_PATH)
	site.GetBool("web server mode", &CORE_WEB_SERVER_MODE)

	site.GetString("social media image", &SITE_SOCIAL_MEDIA_IMAGE)
	SITE_SOCIAL_MEDIA_IMAGE = URL_IMAGE_PATH + SITE_SOCIAL_MEDIA_IMAGE
	site.GetString("default avatar", &DEFAULT_AVATAR)
	DEFAULT_AVATAR = URL_IMAGE_PATH + DEFAULT_AVATAR

	site.GetString("cookie domain", &COOKIE_DOMAIN)
	site.GetString("cookie home path", &COOKIE_HOME_PATH)

	//social media
	site.GetString("official twitter handle", &SITE_TWITTER)
	site.GetBool("enable facebook login", &ENABLE_FACEBOOK_LOGIN)
	site.GetString("facebook app id", &FACEBOOK_APP_ID)
	site.GetString("facebook app secret", &FACEBOOK_APP_SECRET)

	//mail
	site.GetString("smtp server", &SMTP_SERVER)
	site.GetInt("smtp port", &SMTP_PORT)
	site.GetString("email sender address", &SMTP_SENDER_ADDRESS)

	//daemon
	daemon.GetBool("enable setuid", &DAEMON_SETUID)
	daemon.GetInt("user id", &DAEMON_USER_ID)
	daemon.GetInt("group id", &DAEMON_GROUP_ID)

	daemon.GetBool("enable chroot", &DAEMON_CHROOT)
	daemon.GetString("chroot dir", &DAEMON_SITE_DIR)

	//core
	core.GetFloat64("token age", &CORE_TOKEN_AGE)
	core.GetInt("max os threads", &CORE_GO_MAX_PROCS)
	core.GetInt("cache control", &CORE_CACHE_CONTROL)
	core.GetInt("db cache timeout", &CORE_DB_CACHE_TIMEOUT)
	//core.GetUint("user cache timeout", &CORE_USER_CACHE_TIMEOUT)
	//core.GetUint("disk cache timeout", &CORE_DISK_CACHE_TIMEOUT)
	core.GetUint64("disk cache limit", &CORE_DISK_CACHE_LIMIT)
	core.GetInt("disk cache purger", &CORE_DISK_CACHE_PURGER)

	core.GetInt("session cookie age", &CORE_SESSION_AGE)
	core.GetInt("platform preference cookie age", &CORE_PLATFORM_PREF_AGE)
	core.GetBool("use referrer cookie", &CORE_REFERRER_COOKIE)
	core.GetInt("referrer cookie age", &CORE_REFERRER_AGE)
	core.GetInt("username min characters", &CORE_USERNAME_MIN_CHARS)
	core.GetInt("username max characters", &CORE_USERNAME_MAX_CHARS)
	core.GetInt("password min characters", &CORE_PASSWORD_MIN_CHARS)
	core.GetInt("password max characters", &CORE_PASSWORD_MAX_CHARS)
	core.GetInt("email max characters", &CORE_EMAIL_MAX_CHARS)
	core.GetInt("twitter max characters", &CORE_TWITTER_MAX_CHARS)

	core.GetUint("user agent db crop", &CORE_USER_AGENT_CROP)
	core.GetUint("source url crop", &CORE_SOURCE_URL_CROP)
	core.GetBool("registration requires email", &CORE_REG_EMAIL_REQUIRED)
	core.GetBool("registration requires twitter", &CORE_REG_TWITTER_REQUIRED)
	core.GetInt("post min characters", &CORE_POST_MIN_CHARS)
	core.GetInt("post max characters", &CORE_POST_MAX_CHARS)
	core.GetInt("thread title min characters", &CORE_THREAD_TITLE_MIN_C)
	core.GetInt("thread title max characters", &CORE_THREAD_TITLE_MAX_C)
	core.GetInt("thread link url max characters", &CORE_THREAD_LINK_URL_MAX_C)
	core.GetInt("thread link desc min characters", &CORE_THREAD_LINK_DESC_MIN_C)
	core.GetInt("thread link desc max characters", &CORE_THREAD_LINK_DESC_MAX_C)
	core.GetInt("thread link content type max characters", &CORE_THREAD_LINK_CT_MAX_C)
	core.GetUint("max nested comments", &CORE_MAX_COMMENT_NEST)

	// appearance
	core.GetBool("enable executables", &CORE_ENABLE_SYS_EXEC)
	core.GetBool("allow robots", &CORE_ALLOW_ROBOTS)
	core.GetBool("allow registration", &CORE_ALLOW_REGISTRATION)
	core.GetBool("minify desktop html", &CORE_MINIFY_DESKTOP)
	core.GetBool("minify mobile html", &CORE_MINIFY_MOBILE)
	core.GetString("minify js css executable", &CORE_MINIFY_JS_CSS)

	core.GetInt("avatar small width", &CORE_AVATAR_SMALL_WIDTH)
	core.GetInt("avatar small height", &CORE_AVATAR_SMALL_HEIGHT)
	core.GetInt("avatar large width", &CORE_AVATAR_LARGE_WIDTH)
	core.GetInt("avatar large height", &CORE_AVATAR_LARGE_HEIGHT)
	core.GetString("date format article", &DATE_FMT_ARTICLE)
	core.GetString("date format thread", &DATE_FMT_THREAD)

	// websockets
	core.GetBool("websockets desktop", &ENABLE_REAL_TIME_DESKTOP)
	core.GetBool("websockets mobile", &ENABLE_REAL_TIME_MOBILE)

}
