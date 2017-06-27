// admin.go
// provides /root/

package main

import (
	"github.com/lmorg/laurencemorgan.co.uk/level10fireball/gallery"
	"os"
)

var galleryStatus string

func pageRoot(session *Session) string {
	session.Page.Section = &session.Special
	if !session.User.Permissions.Match(PERM_ROOT) {
		pageDenied(session, "page")
		return ""
	}
	switch session.Path[2] {
	case "reloadNode":
		loadEnvironmentFromDB(&live_layout)
		return SITE_HOME_URL

	case "galleryUpdate":
		go gallery.LsImages(session.GetQueryString("path").Value, &galleryStatus)
		session.Page.Content += "Gallery update spawned as background process.<_br|>Please avoid visiting the destination page until process is complete."
		return ""

	case "galleryUpdateStatus":
		session.Page.Content += galleryStatus
		return ""

	case "die":
		os.Exit(1)

	default:
		page404(session, "root management page")
	}

	return ""
}
