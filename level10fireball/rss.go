// rss.go
package main

import ()

const (
	rss_header string = `
<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
`

	rss_footer string = `</rss>`

	rss_item string = `
<item>
 <title>%s</title>
 <description>%s</description>
 <link>%s</link>
 <pubDate>%s</pubDate>
</item>
`
)

func rss(session *Session) {
	//topic_id, _ := Atoui(session.Path[2])
}
