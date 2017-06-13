// parse websites

package main

import (
	enc_json "encoding/json"
	"errors"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var (
	rx_og_head     *regexp.Regexp
	rx_og_title    *regexp.Regexp
	rx_og_metas    *regexp.Regexp
	rx_og_props    *regexp.Regexp
	og_description []string
	og_title       []string
)

type URLParse struct {
	Title        string
	Description  string
	Content_Type string
	Err          error
}

func (r URLParse) ToJSONobj() (json string, err error) {
	j := make(map[string]string)

	j["Title"] = r.Title
	j["Desc"] = r.Description
	j["Mime"] = r.Content_Type
	if r.Err != nil {
		j["Err"] = r.Err.Error()
	}

	var b []byte
	b, err = enc_json.Marshal(j)

	json = string(b)
	return
}

func init() {
	rx_og_head = regexCompile(`(?imsU)<head(\s+.*|)>(.*)</head>`)
	rx_og_title = regexCompile(`(?imsU)<title(\s+.*|)>(.*)</title>`)
	rx_og_metas = regexCompile(`(?imsU)<meta +(.*)([\s]+|)>`)
	rx_og_props = regexCompile(`(?imsU)([\-a-z0-9]+)="(.*)"`)
	og_description = []string{"og:description", "twitter:description", "description"}
	og_title = []string{"og:title", "twitter:title", "title"}
}

func ogGet(http_resp *http.Response) (r URLParse) {
	var (
		http_body []byte
	)

	r.Content_Type = strings.Split(http_resp.Header.Get("content-type"), ";")[0]

	if http_resp.StatusCode != 200 {
		r.Err = errors.New("Document returned with a HTTP status " + http_resp.Status)
		return
	}

	switch r.Content_Type {
	case "text/html":
		http_body, r.Err = ioutil.ReadAll(http_resp.Body)
		if r.Err != nil {
			return
		}
		html_head := rx_og_head.FindAllStringSubmatch(string(http_body), 1)
		if len(html_head) == 0 || len(html_head[0]) < 3 {
			r.Err = errors.New("Unable to reliably extract sample text (malformed <head> tags in HTML document). Please add your own title and description.")
			return
		}
		ogParser(html_head[0][2], &r)
		if r.Title == "" || r.Description == "" {
			r.Err = errors.New("Unable to reliably extract sample text. Please manually populate the empty fields.")
		}
		r.Title = html.UnescapeString(r.Title)
		r.Description = html.UnescapeString(r.Description)
		return

	case "text/plain":
		http_body, r.Err = ioutil.ReadAll(http_resp.Body)
		if r.Err != nil {
			return
		}
		r.Description = trimString(string(http_body), uint(CORE_THREAD_LINK_DESC_MAX_C)).HTMLEscaped()
		r.Err = errors.New("Unable to extract a page title.")
		return

	default:
		r.Err = errors.New("Destination content was not parsed (Content type '" + r.Content_Type + "'). Please add your own title and description.")
	}

	return
}

func ogParser(html_head string, r *URLParse) {
	// code tested on: http://play.golang.org/p/30nCYTlGfS
	meta_tags := make(map[string]string)
	meta := rx_og_metas.FindAllStringSubmatch(html_head, -1)
	for i := 0; i < len(meta); i++ {
		prop := rx_og_props.FindAllStringSubmatch(meta[i][0], -1)
		var name, content string
		for j := 0; j < len(prop); j++ {
			s := strings.ToLower(prop[j][1])
			if s == "name" || s == "property" {
				name = strings.ToLower(prop[j][2])
			} else if strings.ToLower(prop[j][1]) == "content" {
				content = prop[j][2]
			}
		}
		meta_tags[name] = content
	}

	// find description meta headers
	for i := 0; i < len(og_description); i++ {
		r.Description = meta_tags[og_description[i]]
		if r.Description != "" {
			break
		}
	}

	// find title meta headers
	for i := 0; i < len(og_title); i++ {
		r.Title = meta_tags[og_title[i]]
		if r.Title != "" {
			return
		}
	}

	// no title found. Fall back to <title></title> tags
	title := rx_og_title.FindAllStringSubmatch(html_head, 1)
	if len(title) > 0 && len(title[0]) > 2 {
		r.Title = title[0][2]
	}

	return
}
