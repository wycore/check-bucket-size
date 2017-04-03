package main

type ReturnCode int

const (
	OK ReturnCode = 0 + iota
	WARNING
	CRITICAL
	UNKNOWN
)
