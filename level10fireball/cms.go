// cms
package main

import (
	"fmt"
	"github.com/lmorg/laurencemorgan.co.uk/level10fireball/gallery"
	"html"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

var (
	rx_variables *regexp.Regexp

	rx_cms_split    *regexp.Regexp
	rx_html_prefix  *regexp.Regexp
	rx_whitespace   *regexp.Regexp
	rx_purge_tags   *regexp.Regexp
	rx_tags_pre     *regexp.Regexp
	rx_tags_txtarea *regexp.Regexp

	rx_hash_tag *regexp.Regexp
	rx_at_tag   *regexp.Regexp

	// bbcode
	rx_tags_all_bb     *regexp.Regexp
	rx_tags_b          *regexp.Regexp
	rx_tags_b_         *regexp.Regexp
	rx_tags_i          *regexp.Regexp
	rx_tags_i_         *regexp.Regexp
	rx_tags_u          *regexp.Regexp
	rx_tags_u_         *regexp.Regexp
	rx_tags_code       *regexp.Regexp
	rx_tags_code_      *regexp.Regexp
	rx_tags_quote      *regexp.Regexp
	rx_tags_quote_     *regexp.Regexp
	rx_tags_quotes     *regexp.Regexp
	rx_tags_img        *regexp.Regexp
	rx_tags_img_       *regexp.Regexp
	rx_tags_url        *regexp.Regexp
	rx_tags_url_       *regexp.Regexp
	rx_hyperlink       *regexp.Regexp
	rx_newline_no_br1  *regexp.Regexp
	rx_newline_no_br2  *regexp.Regexp
	rx_newline_comment *regexp.Regexp
	rx_newline_chr13   *regexp.Regexp
)

func init() {
	rx_variables = regexCompile(`(?i)\$\{[a-z]+\}`)

	rx_cms_split = regexCompile("(?s)<_.*?>")
	rx_html_prefix = regexCompile(`(?i)^http(s|):\/\/`)
	rx_whitespace = regexCompile(`[\s]+`)
	rx_purge_tags = regexCompile(`(?s)<.*?[^\\]>`)
	//rx_purge_tags = regexCompile(`<.*?>`)
	rx_tags_pre = regexCompile(`(?si)<pre.*?>(.*?)</pre>`)
	rx_tags_txtarea = regexCompile(`(?si)<textarea.*?>(.*?)</textarea>`)

	rx_hash_tag = regexCompile(`(?i)([^0-9A-Za-z_&]|^)(#[0-9a-z]+)`) // TODO: build hash tags routine.
	rx_at_tag = regexCompile(`([^0-9A-Za-z_&]|^)(@[0-9]+)`)

	// bb code
	rx_tags_all_bb = regexCompile(`(?i)\[(\/|)(b|i|u|code|img|url)\]`)

	rx_tags_b = regexCompile(`(?i)\[b\]`)
	rx_tags_b_ = regexCompile(`(?i)\[\/b\]`)
	rx_tags_i = regexCompile(`(?i)\[i\]`)
	rx_tags_i_ = regexCompile(`(?i)\[\/i\]`)
	rx_tags_u = regexCompile(`(?i)\[u\]`)
	rx_tags_u_ = regexCompile(`(?i)\[\/u\]`)

	rx_tags_code = regexCompile(`(?i)\[code\]`)
	rx_tags_code_ = regexCompile(`(?i)\[\/code\]`)
	rx_tags_quote = regexCompile(`(?i)\[quote\]`)
	rx_tags_quote_ = regexCompile(`(?i)\[\/quote\]`)
	rx_tags_quotes = regexCompile(`(?imsU)\[quote\].*\[\/quote\]`)
	rx_tags_img = regexCompile(`(?i)\[img\]`)
	rx_tags_img_ = regexCompile(`(?i)\[\/img\]`)
	rx_tags_url = regexCompile(`(?i)\[url\]`)
	rx_tags_url_ = regexCompile(`(?i)\[\/url\]`)

	//rx_hyperlink = regexCompile(`(?i)((http|https|ftp):\/\/.*?)(\s|\[|\]|\(|\)|<|>|\||$)`)
	rx_hyperlink = regexCompile(`(?i)(^|[^\"])(\s+|)((http|https|ftp):\/\/.*?)(\"|\s|\[|\]|\(|\)|\<|\>|\||$)`)
	rx_newline_no_br1 = regexCompile(`>(\s|)+?\n`)
	rx_newline_no_br2 = regexCompile(`\n(\s|)+?<`)
	rx_newline_comment = regexCompile(`<!--\\n-->`)
	rx_newline_chr13 = regexCompile(`(\r\n|\n|\r)`)
}

func cmsHashTag(content *string) {
	// Test code: http://play.golang.org/p/H3Obq3j0pk
	m := rx_hash_tag.FindAllStringSubmatch(*content, -1)
	for i, _ := range m {
		*content = strings.Replace(*content, m[i][0],
			fmt.Sprintf(`%s<a href="%stag/%s" title="hash tag: &#35;%s">&#35;%s</a>`, m[i][1], SITE_HOME_URL, m[i][2][1:], m[i][2][1:], m[i][2][1:]), 1)
	}
}

func cmsAtTag(content *string) {
	// Test code: http://play.golang.org/p/H3Obq3j0pk
	m := rx_at_tag.FindAllStringSubmatch(*content, -1)
	for i, _ := range m {
		*content = strings.Replace(*content, m[i][0],
			fmt.Sprintf(`%s<a href="%suser/%s" title="user ID: @%s">@%s</a>`, m[i][1], SITE_HOME_URL, m[i][2][1:], m[i][2][1:], m[i][2][1:]), 1)
	}
}

func cmsPlainFormattingS(content string) (plain string) {
	return cmsPlainFormatting(&content)
}

func cmsPlainFormatting(content *string) (plain string) {
	plain = rx_purge_tags.ReplaceAllString(*content, "")
	plain = rx_whitespace.ReplaceAllString(plain, " ")

	return
}

func cmsPurgeBBCode(content *string) (plain string) {
	plain = rx_tags_quotes.ReplaceAllString(*content, "")
	plain = rx_tags_all_bb.ReplaceAllString(plain, "")

	return
}

func cmsMinify(content *string, session *Session) {
	var replace string
	matched := rx_tags_pre.FindAllStringSubmatch(*content, -1)
	if len(matched) > 0 {
		for _, match := range matched {
			//log.Println(i, match)
			replace = strings.Replace(match[1], "\n", "<_special|n>", -1)
			replace = strings.Replace(replace, "\t", "<_special|t>", -1)
			replace = strings.Replace(replace, " ", "<_special|s>", -1)
			*content = strings.Replace(*content, match[1], replace, -1)
		}
	}
	matched = rx_tags_txtarea.FindAllStringSubmatch(*content, -1)
	if len(matched) > 0 {
		for _, match := range matched {
			//log.Println(i, match)
			replace = strings.Replace(match[1], "\n", "<_special|n>", -1)
			replace = strings.Replace(replace, "\t", "<_special|t>", -1)
			replace = strings.Replace(replace, " ", "<_special|s>", -1)
			*content = strings.Replace(*content, match[1], replace, -1)
		}
	}
	*content = rx_whitespace.ReplaceAllString(*content, " ")
	*content = strings.Replace(*content, "<_special|n>", "\n", -1)
	*content = strings.Replace(*content, "<_special|t>", "\t", -1)
	*content = strings.Replace(*content, "<_special|s>", " ", -1)
}

func cmsBBDecode(content *string, session *Session) {
	_code0 := fnTags["code"](&[]string{"", "0", "", "", ""}, session)
	_code1 := fnTags["code"](&[]string{"", "1", "", "", ""}, session)
	_quote0 := fnTags["quote"](&[]string{"", "0", "", "", ""}, session)
	_quote1 := fnTags["quote"](&[]string{"", "1", "qb_comments", "", ""}, session)
	_img0 := `<img class="us_img" src="`
	_img1 := `">`

	*content = html.EscapeString(*content)

	*content = rx_tags_code.ReplaceAllString(*content, _code1)
	*content = rx_tags_code_.ReplaceAllString(*content, _code0)
	*content = rx_tags_quote.ReplaceAllString(*content, _quote1)
	*content = rx_tags_quote_.ReplaceAllString(*content, _quote0)
	*content = rx_tags_img.ReplaceAllString(*content, _img0)
	*content = rx_tags_img_.ReplaceAllString(*content, _img1)

	*content = rx_newline_no_br1.ReplaceAllString(*content, `><!--\n-->`)
	*content = rx_newline_no_br2.ReplaceAllString(*content, `<!--\n--><`)
	*content = rx_newline_chr13.ReplaceAllString(*content, `<br/>`)
	*content = rx_newline_comment.ReplaceAllString(*content, "")

	*content = rx_tags_b.ReplaceAllString(*content, "<b>")
	*content = rx_tags_b_.ReplaceAllString(*content, "</b>")
	*content = rx_tags_i.ReplaceAllString(*content, "<i>")
	*content = rx_tags_i_.ReplaceAllString(*content, "</i>")
	*content = rx_tags_u.ReplaceAllString(*content, "<u>")
	*content = rx_tags_u_.ReplaceAllString(*content, "</u>")

	// purge [url] tags
	// for security reasons, no support planned for [url=http://example.com]text[/url]
	*content = rx_tags_url.ReplaceAllString(*content, "")
	*content = rx_tags_url_.ReplaceAllString(*content, "")

	*content = strings.Replace(*content, ":)", `<span class="smilie">😊</span>`, -1)
	*content = strings.Replace(*content, ":-)", `<span class="smilie">😊</span>`, -1)
	*content = strings.Replace(*content, ":D", `<span class="smilie">😄</span>`, -1)
	*content = strings.Replace(*content, ":-D)", `<span class="smilie">😄</span>`, -1)
	*content = strings.Replace(*content, ";)", `<span class="smilie">😉</span>`, -1)
	*content = strings.Replace(*content, ";-)", `<span class="smilie">😉</span>`, -1)
	*content = strings.Replace(*content, ":P", `<span class="smilie">😋</span>`, -1)
	*content = strings.Replace(*content, ":-P", `<span class="smilie">😋</span>`, -1)
	*content = strings.Replace(*content, ":p", `<span class="smilie">😋</span>`, -1)
	*content = strings.Replace(*content, ":-p", `<span class="smilie">😋</span>`, -1)
	*content = strings.Replace(*content, ":(", `<span class="smilie">😞</span>`, -1)
	*content = strings.Replace(*content, ":-(", `<span class="smilie">😞</span>`, -1)
	*content = strings.Replace(*content, ":-/", `<span class="smilie">😕</span>`, -1)
	*content = strings.Replace(*content, ":&#39;(", `<span class="smilie">😢</span>`, -1)
	*content = strings.Replace(*content, "&lt;3", `<span class="smilie">💛</span>`, -1)

	hyperlinks := rx_hyperlink.FindAllStringSubmatch(*content, -1)
	for i, _ := range hyperlinks {
		*content = strings.Replace(*content, hyperlinks[i][3],
			fmt.Sprintf("%s%s%s",
				fnTags["a"](&[]string{"", hyperlinks[i][3], "user submitted links open in a new window/tab", "_blank", ""}, session),
				hyperlinks[i][3],
				fnTags["a"](&[]string{"", "", "", "", ""}, session)),
			1)
	}

	cmsHashTag(content)
	cmsAtTag(content)
}

func cmsBBNoClose(content *string) (failed string) {
	debugLog(len(rx_tags_b.FindAllString(*content, -1)), len(rx_tags_b_.FindAllString(*content, -1)))

	if len(rx_tags_b.FindAllString(*content, -1)) != len(rx_tags_b_.FindAllString(*content, -1)) {
		return "[b]"
	}

	if len(rx_tags_i.FindAllString(*content, -1)) != len(rx_tags_i_.FindAllString(*content, -1)) {
		return "[i]"
	}

	if len(rx_tags_u.FindAllString(*content, -1)) != len(rx_tags_u_.FindAllString(*content, -1)) {
		return "[u]"
	}

	if len(rx_tags_code.FindAllString(*content, -1)) != len(rx_tags_code_.FindAllString(*content, -1)) {
		return "[code]"
	}

	if len(rx_tags_quote.FindAllString(*content, -1)) != len(rx_tags_quote_.FindAllString(*content, -1)) {
		return "[quote]"
	}

	if len(rx_tags_img.FindAllString(*content, -1)) != len(rx_tags_img_.FindAllString(*content, -1)) {
		return "[img]"
	}

	return ""
}

/*************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************/

/*************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************/

/*************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************/

func cmsLayout(content *string, session *Session) {
	var replace string

	matched := rx_cms_split.FindAllStringSubmatch(*content, -1)

	if len(matched) == 0 {
		return
	}

	for _, match := range matched {
		tag := strings.Split(match[0][2:len(match[0])-1], "|")

		// replace variables
		for i, _ := range tag {
			for key, value := range session.Variables {
				tag[i] = strings.Replace(tag[i], "${"+key+"}", value, -1)
			}
		}

		// first 5 fields are empty strings. It's a kludge but it makes it easier
		// for code later as that code can just assume the array index exists.
		for i := len(tag); i < 5; i++ {
			tag = append(tag, "")
		}
		if fnLayout[tag[0]] != nil {
			replace = fnLayout[tag[0]](&tag, session)
			*content = strings.Replace(*content, match[0], replace, -1)
		}
	}
}

var fnLayout = map[string]func(*[]string, *Session) string{
	"layout": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {

		case "quote", "fortune":
			return session.layout.Quotes[rand.Intn(len(session.layout.Quotes)-1)]

		case "social buttons":
			return session.layout.SocialButtons

		case "switch platform":
			var alt string
			if (*tag)[2] == "desktop" {
				alt = "mobile"
			} else {
				alt = "desktop"
			}
			return fmt.Sprintf(`<span id="switch_platform">You are currently viewing the %s version of the site. <a href="%splatform/%s/">Switch to %s view</a>.</span>`,
				(*tag)[2], SITE_HOME_URL, alt, alt)

		case "menubar":
			var s string

			for i := 0; i < len(session.layout.Menubar); i++ {
				if session.layout.Menubar[i].Label != "<_user session>" {
					s += fmt.Sprintf(`<span class="mb_item">%s<a href="%s" title="%s">%s</a>%s</span>`,
						session.layout.MenubarLeft, session.layout.Menubar[i].URL, session.layout.Menubar[i].Description, session.layout.Menubar[i].Label, session.layout.MenubarRight)
				} else {

					if session.User.ID == 0 {
						s += fmt.Sprintf(`<span class="mb_item">%s<a href="%slogin" title="You are not logged into %s">Login / Register</a>%s</span>`, session.layout.MenubarLeft, SITE_HOME_URL, SITE_NAME, session.layout.MenubarRight)
					} else {
						s += fmt.Sprintf(`<span class="mb_item">%s<a href="%sme" title="You are currently logged in as %s">%s</a>%s</span>`, session.layout.MenubarLeft, SITE_HOME_URL, session.User.Name.Long().Value, session.User.Name.Short().Value, session.layout.MenubarRight)
						s += fmt.Sprintf(`<span class="mb_item">%s<a href="%slogout?t=%s" title="Log out of %s">Logout</a>%s</span>`, session.layout.MenubarLeft, SITE_HOME_URL, session.Token, SITE_NAME, session.layout.MenubarRight)
					}
				}
			}
			return s

		case "breadcrumb":
			s := `<div id="breakcrumb">`
			for i := 2; i < len(*tag); i++ {
				var item string

				switch (*tag)[i] {
				case "home", "site":
					item = fmt.Sprintf(`<div class="bc_item"><a href="%s" title="%s: %s">%s</a></div>`,
						SITE_HOME_URL, SITE_NAME, SITE_DESCRIPTION, mobileBreadcrumbs(session, SITE_NAME).HTMLEscaped())
				case "all topics", "topics":
					if session.Forum.ID > 0 || session.Forum.Title.Value == "root" {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%sforums/" title="%s: Forums">Forums</a></div>`,
							SITE_HOME_URL, SITE_NAME)
					} else {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%stopics/" title="%s: All Topics">All Topics</a></div>`,
							SITE_HOME_URL, SITE_NAME)
					}
				case "topic":
					if session.Forum.ID != 0 || session.Forum.Title.Value == "root" {
						continue
					}
					if session.Topic.ID == 0 {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%slist/" title="%s: List Articles">List Articles</a></div>`,
							SITE_HOME_URL, SITE_NAME)
					} else {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%s" title="%s">%s</a></div>`,
							session.Topic.GetURL("list"), session.Topic.Description.HTMLEscaped(), mobileBreadcrumbs(session, session.Topic.Title.Value).HTMLEscaped())
					}
				case "article":
					if session.Article.ID > 0 {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%s" title="%s">%s</a></div>`,
							session.Article.GetURL("article"), session.Article.Description.HTMLEscaped(), mobileBreadcrumbs(session, session.Article.Title.Value).HTMLEscaped())
					}
				case "all forums", "forums":
					item = fmt.Sprintf(`<div class="bc_item"><a href="%sforums/" title="%s: Forums">Forums</a></div>`,
						SITE_HOME_URL, SITE_NAME)
				case "forum":
					if session.Forum.ID > 0 {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%s" title="%s">%s</a></div>`,
							session.Forum.GetURL("forum"), session.Forum.Description.HTMLEscaped(), mobileBreadcrumbs(session, session.Forum.Title.Value).HTMLEscaped())
					}
				case "thread":
					if session.Thread.ID > 0 {
						item = fmt.Sprintf(`<div class="bc_item"><a href="%s" title="%s">%s</a></div>`,
							session.Thread.GetURL("thread"), session.Thread.Description.HTMLEscaped(), mobileBreadcrumbs(session, session.Thread.Title.Value).HTMLEscaped())
					}
				}
				if item != "" && i > 2 {
					s += session.layout.BreadcrumbSeparator
				}
				s += item
			}
			s += `</div>`
			return s

		case "privacy":
			return session.layout.PrivacyTemplate

			////////////////////////////////////////////////////////////////////
		}
		return ""

		////////////////////////////////////////////////////////////////////////
	},

	// careful here - these can be quite dangerous!
	"querystring": func(tag *[]string, session *Session) string {
		return __form(tag, session)
	},

	"post": func(tag *[]string, session *Session) string {
		return __form(tag, session)
	},

	"token": func(tag *[]string, session *Session) string {
		return session.Token
	},

	/*"user": func(tag *[]string, session *Session) string {
		if (*tag)[1] == "long" {
			return session.User.Name.Long().HTMLEscaped()
		}
		return session.User.Name.Short().HTMLEscaped()
	},*/

	//"userhash": func(tag *[]string, session *Session) string {
	//	return session.User.Hash
	//},

	//"sessionid": func(tag *[]string, session *Session) string {
	//	return session.ID
	//},

	"postprocinc": func(tag *[]string, session *Session) string {
		return session.PostProcInc
	},
}

/*************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************/

/*************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************/

/*************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************
 *************************************************************************************************************************************************************/

func cms(content *string, session *Session) {
	var replace string

	matched := rx_cms_split.FindAllStringSubmatch(*content, -1)

	if len(matched) == 0 {
		return
	}

	for _, match := range matched {
		tag := strings.Split(match[0][2:len(match[0])-1], "|")

		// replace variables
		for i, _ := range tag {
			for key, value := range session.Variables {
				tag[i] = strings.Replace(tag[i], "${"+key+"}", value, -1)
			}
		}

		// first 5 fields are empty strings. It's a kludge but it makes it easier
		// for code later as that code can just assume the array index exists.
		for i := len(tag); i < 5; i++ {
			tag = append(tag, "")
		}
		if fnTags[tag[0]] != nil {
			replace = fnTags[tag[0]](&tag, session)
			*content = strings.Replace(*content, match[0], replace, -1)
		}
	}
}

var fnTags = map[string]func(*[]string, *Session) string{

	"img": func(tag *[]string, session *Session) string {
		return __img(tag, session)
	},

	"img_left": func(tag *[]string, session *Session) string {
		return `<div class="float_left">` + __img(tag, session) + "</div>"
	},

	"img_right": func(tag *[]string, session *Session) string {
		return `<div class="float_right">` + __img(tag, session) + "</div>"
	},

	"gallery": func(tag *[]string, session *Session) string {
		/*
		 *    tag1: template
		 *    tag2: path
		 *    tag3: url
		 */
		session.GalleryID++
		id := fmt.Sprintf("a%dg%d", session.Article.ID, session.GalleryID)
		if session.layout.GalleryTemplates[(*tag)[1]] == (GalleryTemplate{"", ""}) {
			var gt GalleryTemplate
			readTemplate(session, "gallery "+(*tag)[1], &gt.Gallery)

			layout_gallery_template := rx_gthumbs_template.FindString(gt.Gallery)
			gt.Gallery = strings.Replace(gt.Gallery, layout_gallery_template, "<_thumbs>", -1)
			layout_gallery_submatch := rx_gthumbs_template.FindStringSubmatch(layout_gallery_template)
			gt.Thumb = layout_gallery_submatch[2]

			session.layout.GalleryTemplates[(*tag)[1]] = gt
		}
		return gallery.Render(
			session.layout.GalleryTemplates[(*tag)[1]].Gallery,
			session.layout.GalleryTemplates[(*tag)[1]].Thumb,
			(*tag)[2],
			(*tag)[3],
			id)
	},

	"code": func(tag *[]string, session *Session) string {
		/*
		 *       0 disable
		 *       1 enable w/ line numbers
		 *       2> line numbers start @ n
		 *       -1 no line numbers
		 *       $lang = specify language //not implimented
		 */
		//my ($line, $lang) = @_;
		i, err := strconv.Atoi((*tag)[1])
		if i != 0 {
			if !session.Mobile {
				var line_numbers string
				if i > 0 {
					line_numbers = " linenums:" + (*tag)[1]
				}
				return fmt.Sprintf(`</p><pre class="code prettyprint%s">`, line_numbers)

			} else {
				return `</p><pre class="code pre_mobile">`
			}
		}
		if (*tag)[1] == "" || err == nil {
			return "</pre><p>"
		}
		return errWriteHTML(session, err, "CMS parser", `fnTags: "code"`)
	},

	"quote": func(tag *[]string, session *Session) string {
		/*
		 *      1 enable,
		 *      0 disable,
		 *      custom_class // ignored for now
		 */
		if (*tag)[1] != "1" {
			return "</div></div><p>"
		} else {
			//my $class = shift;
			return `
         </p><div class="quoteblock class">
         <div class="quotebegin">&nbsp;</div>
         <div class="quotetext">`
		}
		return ""
	},

	"fixed": func(tag *[]string, session *Session) string {
		// # 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return `<span class="fixed">`
		}
		return "</span>"
	},

	"numbered": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "<ol class=\"list\">"
		}
		return "</ol>"
	},

	"bullets": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "<ol class=\"list\">"
		}
		return "</ol>"
	},

	"o": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "<li>"
		}
		return "</li>"
	},

	"h1": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "</p><h1>"
		}
		return "</h1><p>"
	},

	"h2": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "</p><h2>"
		}
		return "</h2><p>"
	},

	"h3": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "</p><h3>"
		}
		return "</h3><p>"
	},

	"p": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "<p>"
		}
		return "</p>"
	},

	"b": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "<b>"
		}
		return "</b>"
	},

	"i": func(tag *[]string, session *Session) string {
		// 1 enable, 0 disable
		if (*tag)[1] == "1" {
			return "<i>"
		}
		return "</i>"
	},

	"a": func(tag *[]string, session *Session) string {
		// url, title, target
		if (*tag)[1] == "0" || (*tag)[1] == "" {
			return "</a>"
		}
		if (*tag)[3] != "" {
			return fmt.Sprintf(`<a href="%s" title="%s" target="%s">`, (*tag)[1], (*tag)[2], (*tag)[3])
		} else {
			return fmt.Sprintf(`<a href="%s" title="%s">`, (*tag)[1], (*tag)[2])
		}
		return ""
	},

	"br": func(tag *[]string, session *Session) string {
		return "<br/><br/>"
	},

	"url": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {
		case "this":
			return SITE_HOME_URL + session.r.RequestURI[1:]

		case "home":
			return SITE_HOME_URL
		}
		return ""
	},

	"page": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {
		case "title":
			return session.Page.Title().HTMLEscaped()

		case "description":
			if (*tag)[2] != "" {
				i, _ := Atoui((*tag)[2])
				return trimString(session.Page.Description().Value, i).HTMLEscaped()
			}
			return session.Page.Description().HTMLEscaped()

		case "twitter":
			return SITE_TWITTER

		case "image":
			i, err := strconv.Atoi((*tag)[2])
			if err != nil {
				return session.Page.Images[0]
			}
			if i+1 < len(session.Page.Images) && i+1 > 0 {
				return session.Page.Images[i+1]
			}
			return session.Page.Images[0]

		case "og images":
			var s string
			for _, url := range session.Page.Images {
				s += `<meta property="og:image" content="` + url + "\" />\n"
			}
			return s

		}
		return ""
	},

	"thread": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {
		case "title":
			return session.Thread.Title.HTMLEscaped()

		case "description":
			if (*tag)[2] != "" {
				i, _ := Atoui((*tag)[2])
				return trimString(session.Thread.Description.Value, i).HTMLEscaped()
			}
			return session.Thread.Description.HTMLEscaped()
		case "url":
			return fmt.Sprintf("%sthread/%d/%s", SITE_HOME_URL, session.Thread.ID, session.Thread.Title.URLify())
		case "id":
			return Itoa(session.Thread.ID)
		}
		return ""
	},

	"site": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {
		case "name":
			return SITE_NAME

		case "description":
			return SITE_DESCRIPTION

		case "copyright":
			return SITE_COPYRIGHT

		case "version":
			return SITE_VERSION

		case "twitter":
			return SITE_TWITTER

		}
		return ""
	},

	"cdn": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {
		case "images":
			return URL_IMAGE_PATH

		case "layout":
			return URL_STATIC_CONTENT_PATH

		}
		return ""
	},

	"cms": func(tag *[]string, session *Session) string {
		switch (*tag)[1] {
		case "name":
			return CMS_NAME

		case "url":
			return CMS_URL

		case "copyright":
			return CMS_COPYRIGHT

		case "version":
			return version.Version

		case "post max length":
			return strconv.Itoa(CORE_POST_MAX_CHARS) //TODO: do I need this?
		}
		return ""
	},

	"resver": func(tag *[]string, session *Session) string {
		//return fmt.Sprintf("%s-%s", SITE_VERSION, CMS_VERSION)
		return SITE_VERSION + "-" + version.ResString()
	},

	// careful here - this can be quite dangerous!
	"querystring": func(tag *[]string, session *Session) string {
		return __form(tag, session)
	},

	"post": func(tag *[]string, session *Session) string {
		return __form(tag, session)
	},

	"youtube": func(tag *[]string, session *Session) string {
		//#my $id      = shift;
		//#my $title   = shift;
		//my ($id, $title) = @_;
		var (
			size string
		)
		//if !session.Mobile {
		size = fmt.Sprintf(`width="%d" height="%d"`, session.layout.xEmbeddedFrame, session.layout.yEmbeddedFrame)
		//	size = fmt.Sprintf(`width="%d" height="%d"`, 1024, 768)
		//}// else {
		//	size = `width="320" height="240"`
		//}

		//<div class="embedded_container yt_container">
		//<div class="embedded_frame center yt_container"><iframe title="%s" class="youtube-player" type="text/html"
		return fmt.Sprintf(`
       <div class="embedded_container">
         <div class="embedded_frame embedded_yt"><iframe title="%s" type="text/html"
           src="//www.youtube.com/embed/%s" frameborder="0" %s></iframe></div>
         <div class="center image_desc">Youtube video: <a class="image_desc" href="//www.youtube.com/watch?v=%s" target="_blank">%s</a></div>
       </div>`, (*tag)[2], (*tag)[1], size, (*tag)[1], (*tag)[2])
	},

	"fbvideo": func(tag *[]string, session *Session) string {
		// TODO: initialise only gets called once
		initialise := fmt.Sprintf(`
			<div id="fb-root"></div>
			<script>(function(d, s, id) {
			  var js, fjs = d.getElementsByTagName(s)[0];
			  if (d.getElementById(id)) return;
			  js = d.createElement(s); js.id = id;
			  js.src = "//connect.facebook.net/en_GB/sdk.js#xfbml=1&version=v2.4&appId=%s";
			  fjs.parentNode.insertBefore(js, fjs);
			}(document, 'script', 'facebook-jssdk'));</script>`, FACEBOOK_APP_ID)

		html := fmt.Sprintf(`
			<div class="embedded_container">
				<div class="embedded_frame embedded_yt">
					<div class="fb-video" data-href="%s" data-width="%d" data-allowfullscreen="true"></div>
				</div>
				<div class="center image_desc">Facebook video: <a class="image_desc" href="%s" target="_blank">%s</a></div>
			</div>`,
			(*tag)[1],
			session.layout.xEmbeddedFrame,
			(*tag)[1], (*tag)[2])

		return initialise + html
	},

	"googlemap": func(tag *[]string, session *Session) string {
		var src, url, click_tap string
		//my ($src, $title, $mobile) = @_;
		src = strings.Replace((*tag)[1], "source=embed", "output=embed", 1)
		url = strings.Replace((*tag)[1], "output=embed", "source=embed", 1)

		if !session.Mobile {
			click_tap = "Click"
		} else {
			click_tap = "Tap"
		}
		/*if session.Mobile {
			return fmt.Sprintf(`<div><a href="%s" title="%s">%s</a></div>`, url, (*tag)[2],
				__img(&([]string{"", (*tag)[3], (*tag)[2] + " - tap to view a larger map", "", "", ""}), session))
		}*/

		return fmt.Sprintf(`
       <div class="embedded_container">
         <div class="embedded_frame"><iframe title="%s" type="text/html" frameborder="0"
           scrolling="no" marginheight="0" marginwidth="0" width="%d" height="%d" src="%s"></iframe></div>
         <div class="center image_desc"><a href="%s" target="_blank"><span class="nowrap">%s.</span><span class="nowrap">%s here to view a larger map</span></a></div>
       </div>`, (*tag)[2], session.layout.xEmbeddedFrame, session.layout.yEmbeddedFrame, src, url, (*tag)[2], click_tap)
	},

	"miscembed": func(tag *[]string, session *Session) string {
		// my ($src, $owner, $title, $url) = @_;
		return fmt.Sprintf(`
      <div class="center embedded_container">
        <div class="center embedded_frame"><script src="%s"></script></div>
              <span class="center image_desc">%s: <a class="image_desc" href="%s" target="_blank">%s</a></span>
      </div>`, (*tag)[1], (*tag)[2], (*tag)[4], (*tag)[3])
	},

	"sysbin": func(tag *[]string, session *Session) string {
		return html.EscapeString(__execbin(tag, session))
	},

	"cgibin": func(tag *[]string, session *Session) string {
		return __execbin(tag, session)
	},
}

/*
  *
  *  sub _article() {
  *      my $ArticleID = shift;
  *      if (!$ArticleID) { return "</span></a>" }
  *
  *      my @article = &SelectDump("article_summery", "", $ArticleID)
  *                  or return "<span class=\"deadlink\"><a href=\"#\" alt=\"[$lvl10{description}]\" title=\"Dead link\">";
  *
  *      #my $url = &ArticleURL($ArticleID, @article[0]->[0]);
  *      return sprintf "<span><a href=\"%s\" alt=\"%s\" title=\"%s\">",
  *                  &ArticleURL($ArticleID, @article[0]->[0]),
  *                  @article[0]->[0],
  *                  &CMSPreParser(@article[0]->[1],2);
 }
*/

func __img(tag *[]string, session *Session) string {
	if len((*tag)[1]) == 0 {
		return ""
	}
	var (
		url     string = (*tag)[1]
		a_open  string
		a_close string
	)
	if url[:1] == "/" && url[1:2] != "/" {
		url = URL_IMAGE_PATH + url[1:]
	} else if !rx_html_prefix.MatchString(url) {
		url = fmt.Sprintf("%s_img/%s", URL_IMAGE_PATH, url)
	}
	if session.Mobile && ((*tag)[3] == "" || (*tag)[3] == "0") {
		a_open = fmt.Sprintf(`<a href="%s" title="%s">`, url, (*tag)[2])
		a_close = "</a>"
	} else if (*tag)[3] != "" && (*tag)[3] != "0" {
		a_open = fmt.Sprintf(`<a href="%s" title="%s">`, (*tag)[3], (*tag)[2])
		a_close = "</a>"
	}
	session.Page.Images = append(session.Page.Images, url)
	return fmt.Sprintf(`
       <div class="center embedded_container">%s<img src="%s" alt="[%s]" title="%s" class="embedded_image scale_image"/>%s<br/>
       <span class="image_desc">%s</span></div>`, a_open, url, (*tag)[2], (*tag)[2], a_close, strings.Replace((*tag)[2], "\n", "<br/>", -1))
}

func __form(tag *[]string, session *Session) string {
	// read the input
	var value string
	if (*tag)[0] == "querystring" {
		value = session.GetQueryString((*tag)[2]).Value
	} else {
		value = session.GetPost((*tag)[2]).Value
	}

	// check the data type.
	switch (*tag)[1] {
	case "int", "integer":
		i, _ := strconv.Atoi(value)
		value = strconv.Itoa(i)
	case "uint":
		i, _ := Atoui(value)
		value = Itoa(i)
	case "str", "string":
		value = html.EscapeString(value)
	default:
		return "[invalid data type]"
	}

	// store variable or display value?
	if (*tag)[3] != "" {
		debugLog(fmt.Sprintf("Storing %s data '%s' (%s), %s -> %s", (*tag)[0], (*tag)[2], (*tag)[1], (*tag)[3], value))
		session.Variables[(*tag)[3]] = value
		return ""
	}
	return value
}

func __execbin(tag *[]string, session *Session) string {
	// my ($cmd) = @_;
	var params []string = (*tag)[2:]
	for i, _ := range params {
		if params[i] == "" {
			if i == 0 {
				return SystemExecute(session, false, (*tag)[1]).stdout
			}

			params = params[:i]
			break
		}
	}
	return SystemExecute(session, false, (*tag)[1], params...).stdout
}
