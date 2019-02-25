/*
Package uci implements a binding to OpenWRT's UCI (Unified Configuration
Interface) files in pure Go.

The typical use case is reading and modifying UCI config options:
	import "github.com/digineo/go-uci"
	uci.Get("network", "lan", "ifname") //=> []string{"eth0.1"}
	uci.Set("network", "lan", "ipaddr", "192.168.7.1")
	uci.Commit() // or uci.Revert()

For more details head over to the OpenWRT wiki, or dive into UCI's C
source code:
 - https://openwrt.org/docs/guide-user/base-system/uci
 - https://git.openwrt.org/?p=project/uci.git;a=summary

The lexer is heavily inspired by Rob Pike's 2011 GTUG Sydney talk
"Lexical Scanning in Go" (https://talks.golang.org/2011/lex.slide,
https://youtu.be/HxaD_trXwRE), which in turn was a presentation of
an early version of Go's text/template parser. It follows, that this
library borrows code from Go's standard library (BSD-style licensed).

The UCI grammar (for the purpose of this library) is defined as follows:

	uci = configDecl*
	    | (packageDecl configDecl*)+

	packageDecl = "package" value CRLF configDecl* # value:config-name

	configDecl = "config" ident value? CRLF optionDecl* # ident:section-type value:section-name

	optionDecl = "option" ident value # ident:option-name value:option-value
	           | "list" ident value   # ident:option-name value:option-value

	ident = [_a-zA-Z0-9]+

	value = "'" STRING "'"
	      | "\"" STRING "\""
	      | ident
*/
package uci
