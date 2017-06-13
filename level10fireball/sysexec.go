// sysexec.go
package main

import (
	"os/exec"
)

type SysExec struct {
	stdout string
	err    error
}

// TODO: return a pointer instead?
func SystemExecute(session *Session, surpressErrMsg bool, cmd string, params ...string) SysExec {
	if CORE_ENABLE_SYS_EXEC {
		var (
			out []byte
			err error
		)
		out, err = exec.Command(cmd, params...).CombinedOutput()

		if err != nil && err.Error()[:12] != "exit status " {
			isErr(session, err, surpressErrMsg, "system execute request", "SystemExecute")
		}
		return SysExec{string(out), err}
	} else {
		err := raiseErr(session, "SystemExecute is disabled: CORE_ENABLE_SHELL_EXEC = false", surpressErrMsg, "system execute request", "SystemExecute")
		return SysExec{"", err}
	}
}

// TODO: return a pointer instead?
func SystemExecuteSTDOUT(session *Session, surpressErrMsg bool, cmd string, params ...string) SysExec {
	if CORE_ENABLE_SYS_EXEC {
		var (
			out []byte
			err error
		)
		out, err = exec.Command(cmd, params...).Output()

		if err != nil && err.Error()[:12] != "exit status " {
			isErr(session, err, surpressErrMsg, "system execute request", "SystemExecute")
		}
		return SysExec{string(out), err}
	} else {
		err := raiseErr(session, "SystemExecute is disabled: CORE_ENABLE_SHELL_EXEC = false", surpressErrMsg, "system execute request", "SystemExecute")
		return SysExec{"", err}
	}
}

// TODO: replace parent type / function with this one
type SysExecb struct {
	stdout []byte
	err    error
}

func SystemExecuteSTDOUTb(session *Session, surpressErrMsg bool, cmd string, params ...string) (exe *SysExecb) {
	exe = new(SysExecb)
	if CORE_ENABLE_SYS_EXEC {
		exe.stdout, exe.err = exec.Command(cmd, params...).Output()

		if exe.err != nil && exe.err.Error()[:12] != "exit status " {
			isErr(session, exe.err, surpressErrMsg, "system execute request", "SystemExecute")
		}
		return

	} else {
		exe.err = raiseErr(session, "SystemExecute is disabled: CORE_ENABLE_SHELL_EXEC = false", surpressErrMsg, "system execute request", "SystemExecute")
		return
	}
}
