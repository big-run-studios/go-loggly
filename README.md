# go-loggly

This package provides the ability to use both bulk and single log endpoint to ship your logs to loggly.  You must provide a token generated from loggly.

## Installing

### ([godoc](//godoc.org/github.com/parkhubprime/go-loggly))

    go get github.com/parkhubprime/go-loggly

## Setting up your logger
```
package main

import (
	log "github.com/parkhubprime/go-loggly"
)

func main() {
	log.SetupLogger("yourlogglytoken", LogLevelInfo, []string{"test"}, false, true)

	// Print info statement
	Infoln("This is an info statement.")

	// Print info statement with data
	type testStruct struct {
		Name string
		Kind string
	}

	test := &testStruct{
		Name: "Logan",
		Kind: "Log",
	}

	Infod("This is some text", test)

	// Print debug text
	Debugln("This is a debug statement.")

	// Print debug text with additional data.
	Debugf("This is a debug statement %d.", 10000)

	// Print info text
	Infoln("This is an info statement.")

	// Print info text with additional data.
	Infof("This is an info statement %d.", 10000)

	// Print warn text
	Warnln("This is a warning.")

	// Print warn text with additional data.
	Warnf("This is a warning %d.", 10000)

	// Print error text
	Errorln("This is an error.")

	// Print error text with additional data.
	Errorf("This is an error %d.", 10000)

	// Print fatal text
	Fatalln("This is an error.")

	// Print fatal text with additional data.
	Fatalf("This is an error %d.", 10000)
}
```