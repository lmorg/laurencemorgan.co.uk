// functions
package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	rx_urlify1    *regexp.Regexp
	rx_urlify2    *regexp.Regexp
	rx_trim_left  *regexp.Regexp
	rx_trim_right *regexp.Regexp
)

func init() {
	rx_urlify1 = regexCompile(`[^a-z0-9\+\s]`)
	rx_urlify2 = regexCompile(`\s+`)
	rx_trim_left = regexCompile(`^\s+`)
	rx_trim_right = regexCompile(`\s+$`)
}

func trim(s string) string {
	s = rx_trim_left.ReplaceAllString(s, "")
	return rx_trim_right.ReplaceAllString(s, "")
}

func appendS(i uint) string {
	if i == 1 {
		return ""
	}
	return "s"
}

func Atoui(s string) (uint, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	return uint(i), err
}

func Atob(s string) (b byte, err error) {
	i, err := strconv.ParseUint(s, 10, 8)
	return byte(i), err
}

func Itoa(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

/*
func Btoa(i byte) string {
	return strconv.FormatUint(uint64(i), 10)
}
*/
func urlify(s string) string {
	s = strings.ToLower(s)
	s = rx_urlify1.ReplaceAllString(s, "")
	s = rx_urlify2.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	s = strings.TrimSpace(s)
	return s
}

func windowsfyPath(s string) string {
	//debugLog("[functions] [windowsfyPath]", s)
	return s
}

func trimString(s string, i uint) DisplayText {
	// TODO: at some point this will need to be improved to work better with multi-byte unicode
	if uint(len(s)) <= i {
		return DisplayText{s}
	}
	return DisplayText{s[:i-3] + "..."}
}

func trimPString(s *string, i uint) {
	// TODO: at some point this will need to be improved to work better with multi-byte unicode
	if uint(len(*s)) <= i {
		return
	}
	*s = (*s)[:i-3]
	*s += "..."
}

func mobileBreadcrumbs(session *Session, item string) DisplayText {
	if !session.Mobile {
		return DisplayText{item}
	}
	return trimString(item, session.layout.nCharsMobileBreadcrumbs)
}

func regexCompile(s string) *regexp.Regexp {
	r, err := regexp.Compile(s)
	isErr(nil, err, true, "Regex compile", "init")
	return r
}

func dateParse(str_date string, str_func string) time.Time {
	r, err := time.Parse(DATE_FMT_DATABASE, str_date)
	isErr(nil, err, true, "Date conversion", str_func)
	return r
}

func randomString(seed int64, length int) string {
	randInt := func(r *rand.Rand, min int, max int) int {
		return min + r.Intn(max-min)
	}

	r := rand.New(rand.NewSource(seed))
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(randInt(r, 33, 126))
	}
	return string(bytes)
}

func matchTokens(session *Session, source string) bool {
	//debugLog(fmt.Sprintf("matchTokens():\nSource: %s\nToken:  %s\nMatch:  %t", source, session.Token, source == session.Token || source == session.CreateToken(-1)))
	return source == session.Token || source == session.CreateToken(-1)
}

func validationHash(session *Session, key string) string {
	hash := sha512.New()
	_, err := hash.Write([]byte(key + Itoa(session.User.ID) + session.ID + SITE_NAME + session.User.Salt))
	isErr(session, err, false, "creating sha512 hash", "validationHash")
	s := strings.Replace(base64.URLEncoding.EncodeToString(hash.Sum(nil)), "=", ".", -1)
	return s
}

func getReferrer(session *Session) (ref string) {
	// i cheat by putting the referrer in a cookie
	if CORE_REFERRER_COOKIE {
		ref = session.GetCookie("ref").Value

		// but fallback to the HTTP referrer header if level10 is compiled with the ref cookie disabled
	} else {
		ref = session.r.Referer()
	}
	debugLog("ref: ", ref)

	// check the referrer is valid and from our site - otherwise return the site's home URL
	if len(ref) < len(SITE_HOME_URL) || ref[:len(SITE_HOME_URL)] != SITE_HOME_URL {
		return SITE_HOME_URL
	}
	return
}

func getLast(session *Session) (last string) {
	// i cheat by putting the referrer in a cookie
	if CORE_REFERRER_COOKIE {
		last = session.GetCookie("last").Value

		// but fallback to the HTTP referrer header if level10 is compiled with the last cookie disabled
	} else {
		last = session.r.Referer()
	}
	debugLog("last: ", last)

	// check the referrer is valid and from our site - otherwise return the site's home URL
	if len(last) < len(SITE_HOME_URL) || last[:len(SITE_HOME_URL)] != SITE_HOME_URL {
		return SITE_HOME_URL
	}
	return
}

func httpClient(session *Session, url string) (client *http.Client, request *http.Request, err error) {
	client = new(http.Client)

	request, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	request.Header.Add("User-Agent", fmt.Sprintf("%s/%s (%s; +%s)", strings.Replace(CMS_NAME, " ", "", -1), strings.Replace(CMS_VERSION, " ", "", -1), strings.Replace(SITE_NAME, " ", "", -1), SITE_HOME_URL))
	if session.Language != "" {
		request.Header.Add("Accept-Language", session.Language)
	}

	return
}

func httpRequest(client *http.Client, request *http.Request, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	return client.Do(request)

}

func httpGetBinary(session *Session, src, dest string) (err error) {
	// HTTP request
	r, err := httpRequest(httpClient(session, src))
	if err != nil {
		isErr(session, err, true, fmt.Sprintf("cannot GET %s: %s", src, err.Error()), "getBinary")
		return
	}

	// download
	b, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		isErr(session, err, true, "cannot ioutil.ReadAll(r.Body)", "getBinary")
		return
	}

	// write to disk
	ioutil.WriteFile(dest, b, 0644)
	isErr(session, err, true, fmt.Sprintf("writing to %s", dest), "getBinary")

	return
}

func resizeAvatar(session *Session, file io.Reader) (err error) {
	var (
		img image.Image
		//conf      image.Config
		f_small   *os.File
		f_large   *os.File
		img_small image.Image
		img_large image.Image
	)

	/*// get image dimentions
	if conf, err = jpeg.DecodeConfig(file); err != nil {
		isErr(session, err, true, "jpeg.DecodeConfig(file)" , "resizeAvatar")
		return
	}*/

	// decode jpeg into image.Image
	if img, err = jpeg.Decode(file); err != nil {
		isErr(session, err, true, "jpeg.Decode(file)", "resizeAvatar")
		return
	}

	// resize to width / hight
	// and preserve aspect ratio
	//if conf.Width > conf.Height {
	img_large = resize.Resize(0, uint(CORE_AVATAR_LARGE_WIDTH), img, resize.Bicubic)
	img_small = resize.Resize(0, uint(CORE_AVATAR_SMALL_WIDTH), img, resize.Bicubic)
	/*} else {
		img_large = resize.Resize(uint(CORE_AVATAR_LARGE_HEIGHT), 0, img, resize.Bicubic)
		img_small = resize.Resize(uint(CORE_AVATAR_SMALL_HEIGHT), 0, img, resize.Bicubic)
	}*/

	if f_small, err = os.Create(fmt.Sprintf("%savatars/%d-small.jpg", PWD_WRITABLE_PATH, session.User.ID)); err != nil {
		isErr(session, err, true, fmt.Sprintf("%savatars/%d-small.jpg", PWD_WRITABLE_PATH, session.User.ID), "resizeAvatar")
		return
	}
	if f_large, err = os.Create(fmt.Sprintf("%savatars/%d-large.jpg", PWD_WRITABLE_PATH, session.User.ID)); err != nil {
		isErr(session, err, true, fmt.Sprintf("%savatars/%d-large.jpg", PWD_WRITABLE_PATH, session.User.ID), "resizeAvatar")
		return
	}

	jpeg.Encode(f_small, img_small, nil)
	jpeg.Encode(f_large, img_large, nil)

	f_small.Close()
	f_large.Close()
	return

}
