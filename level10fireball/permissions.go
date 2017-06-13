// Stock permissions:

package main

const (
	PERM_EVERYONE_READ string = ""  // read access to everyone. write access still requires registration.
	PERM_NOBODY        string = "x" // only used against core records - not recommended for normal use.
	PERM_UNREGISTERED  string = "0" // unregistered users only.
	PERM_REGISTERED    string = "1" // registered (default for users).
	PERM_FORUM_MOD     string = "2" // forum moderators.
	PERM_FORUM_ADMIN   string = "3" // forum administrators.
	PERM_CMS_WRITER    string = "4" // Article / blog item writers.
	PERM_CMS_EDITOR    string = "5" // Article / blog item editors.
	PERM_ROOT          string = "6" // root - site sysadmins.
)
