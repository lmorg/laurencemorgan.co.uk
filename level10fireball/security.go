// security
package main

import (
	"fmt"
	"os"
	"syscall"
)

func secureDaemon() {
	if DAEMON_CHROOT {
		// WARNING: this may break other things. Use at your own risk
		infoLog(fmt.Sprintf("Chrooting to %s....", DAEMON_SITE_DIR))
		failOnErr(os.Chdir(DAEMON_SITE_DIR), "secureDaemon->os.Chdir")
		failOnErr(syscall.Chroot(DAEMON_SITE_DIR), "secureDaemon->syscall.Chroot")
	}

	if DAEMON_SETUID {
		infoLog(fmt.Sprintf("Changing daemon UID:GID (%d:%d)....", DAEMON_USER_ID, DAEMON_GROUP_ID))
		failOnErr(syscall.Setgid(DAEMON_GROUP_ID), "secureDaemon->syscall.Setgid")
		failOnErr(syscall.Setuid(DAEMON_USER_ID), "secureDaemon->syscall.Setuid")
	}
}
