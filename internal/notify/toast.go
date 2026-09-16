// Package notify pushes OS-level toast notifications. It wraps go-toast,
// which was already an indirect dependency of the Wails module before this
// package existed (pulled in transitively, unused by our own code) — see
// CLAUDE.md's now-resolved "OS-level push notifications" open question.
package notify

import "git.sr.ht/~jackmordaunt/go-toast/v2"

const appID = "FocusUp"

// Push shows a Windows toast notification with the given title and body.
// On non-Windows platforms the underlying call is a no-op.
func Push(title, body string) error {
	n := toast.Notification{
		AppID: appID,
		Title: title,
		Body:  body,
	}

	return n.Push()
}
