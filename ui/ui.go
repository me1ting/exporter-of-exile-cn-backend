package ui

import "github.com/lxn/walk"

func ShowError(err error, owner walk.Form) bool {
	if err == nil {
		return false
	}

	showErrorCustom(owner, "错误", err.Error())

	return true
}

func showErrorCustom(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconError)
}

func showWarningCustom(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconWarning)
}
