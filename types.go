package main

// A UNIX executable return code
// 0 = ok, everything else is not
type ReturnCode int

// Possible ReturnCode values for a nagios/icinga/sensu check
// 0 is OK, 1 is WARNING, 2 is CRITICAL, 3 is UNKNOWN
const (
	// OK is value 0 - which means OK
	OK ReturnCode = 0 + iota
	// WARNING is value 1 - which is a WARNING
	WARNING
	// CRITICAL is value 2 - which is a CRITICAL
	CRITICAL
	// UNKNOWN is value 3 - which is UNKNOWN
	UNKNOWN
)
