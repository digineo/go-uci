/*
Package uci implements a binding to OpenWrt's UCI (Unified Configuration
Interface) files in pure Go.

The typical use case is reading and modifying UCI config options:

	import "github.com/digineo/go-uci"

	uci.Get("network", "lan", "ifname") //=> []string{"eth0.1"}, true
	uci.SetType("network", "lan", uci.TypeOption, "ipaddr", "192.168.7.1")
	uci.Commit() // or uci.Revert()

For more details head over to the OpenWrt wiki, or dive into UCI's C
source code:
  - https://openwrt.org/docs/guide-user/base-system/uci
  - https://git.openwrt.org/?p=project/uci.git;a=summary

The lexer is heavily inspired by Rob Pike's 2011 GTUG Sydney talk
"Lexical Scanning in Go" (https://talks.golang.org/2011/lex.slide,
https://youtu.be/HxaD_trXwRE), which in turn was a presentation of
an early version of Go's text/template parser. It follows, that this
library borrows code from Go's standard library (BSD-style licensed).

The UCI grammar (for the purpose of this library) is defined as follows:

	uci
			packageDecl*
			configDecl*

	packageDecl
			`package` value CRLF configDecl*

	configDecl
			`config` ident value? CRLF optionDecl*

	optionDecl
			`option` ident value
			`list` ident value

	ident
			[_a-zA-Z0-9]+

	value
			part+

	part
			`'` [^']* `'`
			`"` ([^"\] | `\` ANY)* `"`
			([^'"\ \t\n#] | `\` ANY)+

Like in libuci, adjacent parts are concatenated without separator.
Single-quoted parts are taken literally, while a backslash escapes the
next character in double-quoted and unquoted parts (a backslash before a
line break continues the value on the next line). When writing configs,
values are single-quoted, and contained single quotes are written as
end quote, escaped quote, start quote:

	option name 'it'\''s'

For now, UCI imports/exports (packageDecl production) are not supported
yet.
*/
package uci
