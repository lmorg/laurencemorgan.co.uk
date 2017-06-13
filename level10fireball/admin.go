// admin.go
// provides /root/

package main

func pageRoot(session *Session) string {
	session.Page.Section = &session.Special
	if !session.User.Permissions.Match(PERM_ROOT) {
		pageDenied(session, "page")
		return ""
	}
	switch session.Path[2] {
	case "reloadnode":
		loadEnvironmentFromDB(&live_layout)
		return SITE_HOME_URL
	default:
		page404(session, "root management page")
	}

	return ""
}
