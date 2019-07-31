/*
Copyright (c) 2014-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

package undolr

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
)

func TestVersionString(t *testing.T) {
	version := GetVersionString()
	if len(version) < 1 {
		t.Errorf("No version string returned")
	}
	if testing.Verbose() {
		t.Log("Testing version: ", version)
	}
}

func TestStart(t *testing.T) {
	err := Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	// Tidy up.
	err = StopAndDiscard()
	if err != nil {
		t.Fatal("Stop:", err)
	}
}

func TestSimpleRecording(t *testing.T) {
	filename, err := tmpnam("")

	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	err = Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	err = Save(filename)
	if err != nil {
		t.Fatal("Save:", err)
	}

	err = StopAndDiscard()
	if err != nil {
		t.Fatal("Stop:", err)
	}

	verifyRecording(t, filename)
}

func TestAsyncRecording(t *testing.T) {
	filename, err := tmpnam("")

	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	err = Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	context, err := Stop()
	if err != nil {
		t.Fatal("Stop:", err)
	}

	err = context.SaveAsync(filename)
	if err != nil {
		t.Fatal("Save:", err)
	}

	previousProgress := 0

	for {
		complete, progress, _, err := context.Poll()
		if err != nil {
			t.Fatal("Poll complete:", err)
		}

		if complete {
			break
		}
		if progress >= 0 {
			if progress < previousProgress {
				t.Fatalf("Progress went backwards (%d -> %d)",
					previousProgress, progress)
			}
			previousProgress = progress
		}
		// Short time interval to actually get some steps
		time.Sleep(10 * time.Millisecond)
	}

	err = context.Discard()
	if err != nil {
		t.Fatal("Discard:", err)
	}

	verifyRecording(t, filename)
}

func TestAsyncRecordingSelect(t *testing.T) {
	filename, err := tmpnam("")

	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	err = Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	context, err := Stop()
	if err != nil {
		t.Fatal("Stop:", err)
	}

	err = context.SaveAsync(filename)
	if err != nil {
		t.Fatal("Save:", err)
	}

	fd, err := context.GetSelectDescriptor()
	if err != nil {
		t.Fatal("GetSelectDescriptor:", err)
	}

	// Read from the FD
	data := make([]byte, 1, 1)
	n, err := syscall.Read(fd, data)
	if n != 1 {
		t.Fatal("Read failed:", err)
	}

	err = context.Discard()
	if err != nil {
		t.Fatal("Discard:", err)
	}

	verifyRecording(t, filename)
}

func TestAsyncRecordingSaveBackground(t *testing.T) {
	err := Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	context, err := Stop()
	if err != nil {
		t.Fatal("Stop:", err)
	}

	filename, err := tmpnam("")

	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	ch := make(chan error, 1)

	go context.SaveBackground(filename, ch)
	select {
	case err = <-ch:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second * 30):
		t.Fatal("Save hadn't completed after 30 seconds")
	}

	err = context.Discard()
	if err != nil {
		t.Fatal("Discard:", err)
	}

	verifyRecording(t, filename)
}

func TestAsyncRecordingIgnored(t *testing.T) {
	err := Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	context, err := Stop()
	if err != nil {
		t.Fatal("Stop:", err)
	}

	if context == nil {
		t.Fatal("No context")
	}

	// Wrap the finalizer for the context with a function:
	//  - checks the finalizer panics
	//  - sends to a channel
	ch := make(chan bool, 1)

	runtime.SetFinalizer(context, nil)
	runtime.SetFinalizer(context, func(c *RecordingContext) {
		defer func() {
			ch <- true
			if r := recover(); r == nil {
				t.Fatalf("Finalizer didn't panic")
			}
		}()
		recordingContextFinalizer(c)
	})

	// We don't call Discard before the context returned by Stop
	// disappears. This should trigger the Finalizer.
	context = nil
	runtime.GC()

	select {
	case <-ch:
	case <-time.After(time.Second * 10):
		t.Fatalf("Finalizer hadn't run after 10 seconds")
	}
}

func doTestSaveOnTermination(t *testing.T, cancel bool) {
	filename, err := tmpnam("")

	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	// We need to launch a separate process that will actually terminate
	// to test save on termination. We abuse TestSaveOnTerminationCancel
	// to avoid having a helper test that shows up in normal runs.
	args := []string{"-test.run=TestSaveOnTerminationCancel", "--", filename}
	if cancel {
		args = append(args, "cancel")
	}

	cmd := exec.Command(os.Args[0], args...)

	env := os.Environ()
	env = append(env, "GO_TEST_SAVE_HELPER=1")
	cmd.Env = env

	err = cmd.Run()
	if err != nil {
		t.Fatal("Spawn:", err)
	}

	if !cancel {
		verifyRecording(t, filename)
	} else {
		verifyEmptyRecording(t, filename)
	}
}

// This is spawned by TestSaveOnTermination[Cancel] so we have a
// process that actually finishes.
func doTestSaveOnTerminationHelper(t *testing.T) {
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	if len(args) < 1 {
		os.Exit(1)
	}

	err := SaveOnTermination(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "SaveOnTermination %s\n", err)
		os.Exit(2)
	}

	err = Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Start %s\n", err)
		os.Exit(3)
	}

	if len(args) > 1 && args[1] == "cancel" {
		err = SaveOnTerminationCancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "SaveOnTerminationCancel %s\n", err)
			os.Exit(4)
		}
	}

	os.Exit(0)
}

func TestSaveOnTermination(t *testing.T) {
	doTestSaveOnTermination(t, false)
}

// We re-use TestSaveOnTerminationCancel as a hook for running the helper
// that actually uses the SaveOnTermination functions. We use this rather
// than TestSaveOnTermination to simplify the test.run argument.
func TestSaveOnTerminationCancel(t *testing.T) {
	if os.Getenv("GO_TEST_SAVE_HELPER") == "1" {
		doTestSaveOnTerminationHelper(t)
	} else {
		doTestSaveOnTermination(t, true)
	}
}

func TestEventLogSize(t *testing.T) {
	size, err := EventLogSizeGet()
	if err != nil {
		t.Fatal("EventLogSizeGet:", err)
	}

	size *= 2
	err = EventLogSizeSet(size)
	if err != nil {
		t.Fatal("EventLogSizeSet:", err)
	}

	newSize, err := EventLogSizeGet()
	if err != nil {
		t.Fatal("EventLogSizeGet:", err)
	}

	if newSize != size {
		t.Fatalf("Size doesn't match (%d vs %d)\n", size, newSize)
	}

	size /= 2
	err = EventLogSizeSet(size)
	if err != nil {
		t.Fatal("EventLogSizeSet:", err)
	}
}

func TestIncludeSymbolFiles(t *testing.T) {
	filenameWith, err := tmpnam("")
	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filenameWith)

	filenameWithout, err := tmpnam("")
	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filenameWithout)

	err = Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	err = Save(filenameWith)
	if err != nil {
		t.Fatal("Save:", err)
	}

	err = IncludeSymbolFiles(false)
	if err != nil {
		t.Fatal("IncludeSymbolFiles:", err)
	}

	err = Save(filenameWithout)
	if err != nil {
		t.Fatal("Save:", err)
	}

	err = StopAndDiscard()
	if err != nil {
		t.Fatal("Stop:", err)
	}

	// Restore to default
	IncludeSymbolFiles(true)

	// This test assumes that symbols exist, and that
	// the amount of recording between the two snapshots
	// above will take less space than the symbols.
	sizeWithout, _ := fileSize(filenameWithout)
	sizeWith, _ := fileSize(filenameWith)
	if sizeWithout >= sizeWith {
		t.Fatalf("Filesize without symbols isn't smaller: %d vs %d\n",
			sizeWithout, sizeWith)
	}
}

func TestShmemLogFilename(t *testing.T) {
	filename, err := tmpnam("shmem")
	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	err = ShmemLogFilenameSet(filename)
	if err != nil {
		t.Fatal("ShmemLogFilenameSet:", err, filename)
	}

	getFilename, err := ShmemLogFilenameGet()
	if err != nil {
		t.Fatal("ShmemLogFilenameSet:", err, filename)
	} else if getFilename != filename {
		t.Fatalf("ShmemLogFilenameGet doesn't match (%s vs %s)\n",
			filename, getFilename)
	}

	err = ShmemLogFilenameClear()
	if err != nil {
		t.Fatal("ShmemLogFilenameClear:", err)
	}
}

func TestShmemLogFilenameSetInvalid(t *testing.T) {
	filename, err := tmpnam("shamem")
	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(filename)

	err = ShmemLogFilenameSet(filename)
	if err == nil {
		t.Fatal("Unexpected success with invalid shmem log filename")
	} else if err != syscall.EINVAL {
		t.Fatal("ShmemLogFilenameSet:", err)
	}
}

func TestShmemLogSize(t *testing.T) {
	size, err := ShmemLogSizeGet()
	if err != nil {
		t.Fatal("ShmemLogSizeGet:", err)
	} else if size != 0 {
		t.Fatal("Expected zero size shmem log when not configured, got", size)
	}

	size = 16777216
	err = ShmemLogSizeSet(size)
	if err != nil {
		t.Fatal("ShmemLogSizeSet:", err)
	}

	newSize, err := ShmemLogSizeGet()
	if err != nil {
		t.Fatal("ShmemLogSizeGet:", err)
	}

	if newSize != size {
		t.Fatalf("Size doesn't match (%d vs %d)\n", size, newSize)
	}

	// Restore
	ShmemLogSizeSet(0)
}

func TestShmemLogRecording(t *testing.T) {
	shmemFilename, err := tmpnam("shmem")
	if err != nil {
		t.Fatal("Filename:", err)
	}
	// Shmem recordings require the file to not exist,
	// so we remove it here. We also defer a removal for
	// after the test.
	os.Remove(shmemFilename)
	defer os.Remove(shmemFilename)

	recFilename, err := tmpnam("")
	if err != nil {
		t.Fatal("Filename:", err)
	}
	defer os.Remove(recFilename)

	err = ShmemLogFilenameSet(shmemFilename)
	if err != nil {
		t.Fatal("ShmemLogFilenameSet:", err, shmemFilename)
	}

	err = ShmemLogSizeSet(16777216)
	if err != nil {
		t.Fatal("ShmemLogSizeSet:", err)
	}

	err = Start()
	if err != nil {
		t.Fatal("Start:", err)
	}

	err = Save(recFilename)
	if err != nil {
		t.Fatal("Save:", err)
	}

	// Restore
	StopAndDiscard()
	ShmemLogFilenameClear()
	ShmemLogSizeSet(0)

	verifyShmemRecording(t, shmemFilename)
}

func fileSize(filename string) (size int64, err error) {
	fileinfo, err := os.Stat(filename)
	if err != nil {
		return
	}
	size = fileinfo.Size()
	return
}

func verifyRecording(t *testing.T, filename string) (err error) {
	return verifyHeader(t, filename, []byte("HD\x10\x00\x00\x00UndoDB recording"))
}

func verifyShmemRecording(t *testing.T, filename string) (err error) {
	return verifyHeader(t, filename, []byte("UndoDB shmem log"))
}

func verifyHeader(t *testing.T, filename string, header []byte) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open %s (%s)", filename, err)
	}

	buf := make([]byte, len(header))
	count, err := file.Read(buf)
	if err != nil || count < len(header) {
		t.Fatalf("Failed to read header from %s (%s)", filename, err)
	}

	if !bytes.Equal(header, buf) {
		t.Fatalf("Header not as expected:\n %q\n vs\n %q", header, buf)
	}

	return nil
}

func verifyEmptyRecording(t *testing.T, filename string) (err error) {
	size, err := fileSize(filename)
	if err != nil {
		return
	}

	if size != 0 {
		t.Fatal("Recording not empty:", size)
	}
	return
}

func tmpnam(extension string) (filename string, err error) {
	iterations := 0

Restart:
	tempfile, err := ioutil.TempFile("", "undolr_test_")
	if err != nil {
		return "", err
	}

	name := tempfile.Name()
	tempfile.Close()
	if extension == "" {
		return name, nil
	}

	// Go 1.11 onwards allows use of a * within the TempFile pattern to
	// specify where the random component goes. But we aren't necessarily
	// running on that (Ubuntu 18.04 LTS ships 1.10) so we do this instead.
	os.Remove(name)
	name = name + "." + extension
	tempfile, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if os.IsExist(err) && iterations < 10 {
		iterations++
		goto Restart
	}

	if err != nil {
		return "", err
	}
	tempfile.Close()
	return tempfile.Name(), nil
}
