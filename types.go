package main

// A UNIX executable return code
// 0 = ok, everything else is not
type ReturnCode int

// Possible ReturnCode values for a nagios/icinga/sensu check
// 0 is OK, 1 is WARNING, 2 is CRITICAL, 3 is UNKNOWN
const (
	OK ReturnCode = 0 + iota
	WARNING
	CRITICAL
	UNKNOWN
)
