// admin.go
// provides /root/

package main

import "os"

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

	case "die":
		os.Exit(1)

	default:
		page404(session, "root management page")
	}

	return ""
}
