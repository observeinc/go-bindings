/*
Copyright (c) 2014-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

package undolr

import (
	"fmt"
	"time"
)

func ExampleSaveBackground() {
	// Start a recording
	err := Start()
	if err != nil {
		panic(err)
	}

	// Stop the recording process, defering Discard to clean up
	recContext, err := Stop()
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

func ExamplePoll() {
	// Start a recording
	err := Start()
	if err != nil {
		panic(err)
	}

	// Stop the recording process, defering Discard to clean up
	recContext, err := Stop()
	if err != nil {
		panic(err)
	}
	defer recContext.Discard()

	err = recContext.SaveAsync("recording.undolr")
	if err != nil {
		panic(err)
	}

	// We now have a recording context and we can save it as we please.
	// Use the progress checking polling method.
	fmt.Printf("Saving...\n")
	for {
		complete, progress, _, err := recContext.Poll()
		if err != nil {
			panic(err)
		}

		if complete {
			fmt.Printf("\rComplete            \n")
			break
		}

		if progress >= 0 {
			fmt.Printf("\r% 3d %% complete", progress)
		}
		// Short time interval to actually get some steps
		time.Sleep(20 * time.Millisecond)
	}
}
