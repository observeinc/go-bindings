/*
Copyright (c) 2014-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

// Package undodb provides examples for using other UndoDB packages
package undodb

import (
	"github.com/undoio/go-bindings/undoex"
	"github.com/undoio/go-bindings/undolr"
	"time"
)

func Example_Annotation() {
	// Start a recording
	err := undolr.Start()
	if err != nil {
		panic(err)
	}

	// Add an annotation in the recording
	err = undoex.AnnotationAddInt("Example", "example annotation", 42)
	if err != nil {
		panic(err)
	}

	// Stop the recording process, defering Discard to clean up
	recContext, err := undolr.Stop()
	if err != nil {
		panic(err)
	}
	defer recContext.Discard()

	// We now have a recording context and we can save it as we please.
	// We save it via a goroutine here, so we could do other things while
	// the save continues in the background if we wanted.
	ch := make(chan error, 1)

	go recContext.SaveBackground("recording.undolr", ch)

	select {
	case err = <-ch:
		if err != nil {
			panic(err)
		}
	case <-time.After(time.Second * 30):
		panic("Save hadn't completed after 30 seconds")
	}
}
