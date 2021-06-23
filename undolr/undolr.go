/*
Copyright (c) 2014-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

// Package undolr provides control of the Undo Live Recorder.
//
// This allows an application to create an Undo Recording of itself running,
// which can then be opened using the Undo Debugger (UndoDB).
package undolr

// #include <undolr.h>
// #include <stdlib.h>
// #include <errno.h>
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var lock sync.Mutex

// A RecordingContext provides access to a recording after recording has been stopped.
type RecordingContext struct {
	ctx    C.undolr_recording_context_t
	valid  bool
	saving bool
	file   string
	line   int
}

// A set of error codes returned by methods handling recording contexts.
var (
	ErrRecordingContextStopFailed     = errors.New("stop failed to create recording context")
	ErrRecordingContextDiscarded      = errors.New("recording context already discarded")
	ErrRecordingContextSaveNotStarted = errors.New("saving not yet started")
	ErrSaveBackgroundReadFailed       = errors.New("failed to read when waiting for save")
)

type undoLrError struct {
	code  C.undolr_error_t
	text  string
	errno error
	rc    int
}

func (e undoLrError) Error() string {
	return fmt.Sprintf("%v; %s", e.errno, e.text)
}

func undoLrErrorWrap(rc int, errno error, code C.undolr_error_t) error {
	if code == 0 && rc < 0 {
		return syscall.Errno(-rc)
	}

	wrapped := undoLrError{
		code:  code,
		errno: errno,
		rc:    rc,
	}

	switch code {
	case C.undolr_error_NO_ATTACH_YAMA:
		wrapped.text = "Failure to attach to the application process due to /proc/sys/kernel/yama/ptrace_scope."
	case C.undolr_error_CANNOT_ATTACH:
		wrapped.text = "Failure to attach to the application process."
	case C.undolr_error_LIBRARY_SEARCH_FAILED:
		wrapped.text = "Failed to find dynamic libraries used by application."
	case C.undolr_error_CANNOT_RECORD:
		wrapped.text = "Recording error."
	case C.undolr_error_NO_THREAD_INFO:
		wrapped.text = "Live Recorder was unable to find information about threads."
	case C.undolr_error_PKEYS_IN_USE:
		wrapped.text = "Use of Protection Keys was detected. This is not yet supported."
	default:
		wrapped.text = "Unknown error"
	}

	return wrapped
}

// Start recording the process.
//
// Attaches Live Recorder to the process, and starts recording it.
//
// The process must not already be being recorded, i.e. <Stop>
// must have been called since any previous call to <Start>.
func Start() error {
	var undoError C.undolr_error_t

	lock.Lock()
	defer lock.Unlock()

	rc, errno := C.undolr_start(&undoError)
	if rc != 0 {
		return undoLrErrorWrap(int(rc), errno, undoError)
	}

	return nil
}

// GetVersionString returns the version string for the underlying UndoLR library.
func GetVersionString() string {
	lock.Lock()
	defer lock.Unlock()
	return C.GoString(C.undolr_get_version_string())
}

// Stop recording the process, keeping it for later saving.
//
// Detaches Live Recorder from the process. The program state recorded
// so far is held in memory until a call to Discard on the
// returned context.
//
// The returned RecordingContext must be later freed using Discard.
func Stop() (context *RecordingContext, err error) {
	var rc C.int

	context = &RecordingContext{}

	lock.Lock()
	defer lock.Unlock()

	rc, err = C.undolr_stop(&context.ctx)
	if rc == 0 {
		context.valid = true
		_, context.file, context.line, _ = runtime.Caller(1)
		runtime.SetFinalizer(context, recordingContextFinalizer)
		err = nil
	} else {
		context = nil
		err = ErrRecordingContextStopFailed
	}
	return
}

func recordingContextFinalizer(context *RecordingContext) {
	if context.valid {
		lock.Lock()
		defer lock.Unlock()
		C.undolr_discard(context.ctx)
		panic(fmt.Sprintf("%s:%d: RecordingContext has not been Discarded",
			context.file, context.line))
	}
}

// StopAndDiscard stops the recording and immediately discards it.
func StopAndDiscard() (err error) {
	lock.Lock()
	defer lock.Unlock()
	rc, err := C.undolr_stop((*C.undolr_recording_context_t)(nil))
	if rc == 0 {
		err = nil
	}
	return
}

// Save recorded program history to a named recording file.
//
// Recording state that is currently held in memory is written to the named
// file as a recording loadable by UndoDB.
//
// The caller must be being recorded, ie Start must have been
// successfully invoked without a subsequent call to Stop. To save
// program state when recording is stopped, use SaveAsync.
//
// On return, the full recording file has been written. Therefore depending
// on the amount of data involved this call may take significant time to
// complete. All threads in the calling process will be stopped during this
// time. See also Stop and SaveAsync.
//
// Save may be called any number of times until Stop is called. Each
// subsequent call to Save will contain later execution history,
// but may also overlap with previous recordings depending on the
// size of the event log and how long the caller runs between calls.
func Save(filename string) (err error) {
	cstring := C.CString(filename)
	defer C.free(unsafe.Pointer(cstring))

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_save(cstring)

	if rc != 0 {
		return
	}
	return nil
}

// SaveAsync will save recorded program history to a named recording file.
//
// Recording state that is currently held in memory (but which is no longer
// being appended to) is asynchronously written to the named file as a
// recording loadable by UndoDB.
//
// The save of the recording is asynchronous, but the same recording context
// may not be operated on until the recording has been fully saved.
func (context *RecordingContext) SaveAsync(filename string) (err error) {
	if !context.valid {
		return ErrRecordingContextDiscarded
	}

	cstring := C.CString(filename)
	defer C.free(unsafe.Pointer(cstring))

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_save_async(context.ctx, cstring)
	if rc != 0 {
		return
	}
	context.saving = true
	return nil
}

// Poll reports the status of the current SaveAsync operation.
func (context *RecordingContext) Poll() (complete bool, progress int, result int, err error) {
	if !context.valid {
		err = ErrRecordingContextDiscarded
		return
	}
	if !context.saving {
		err = ErrRecordingContextSaveNotStarted
		return
	}

	var cComplete, cProgress, cResult C.int

	lock.Lock()
	defer lock.Unlock()
	rc, err := C.undolr_poll_saving_progress(context.ctx, &cComplete, &cProgress, &cResult)

	if rc != 0 {
		return
	}

	complete = cComplete != 0
	progress = int(cProgress)
	result = int(cResult)
	err = nil

	return
}

// GetSelectDescriptor retrieves a selectable file descriptor to detect save state changes.
//
// When the associated save is complete a byte is written to the descriptor, allowing it to
// be selected for read to wake up a thread.
//
// The file descriptor is closed and therefore becomes invalid when the corresponding
// recording context is freed up via Discard.
func (context *RecordingContext) GetSelectDescriptor() (fd int, err error) {
	if !context.valid {
		err = ErrRecordingContextDiscarded
		return
	}

	var cFd C.int

	lock.Lock()
	defer lock.Unlock()
	rc, err := C.undolr_get_select_descriptor(context.ctx, &cFd)
	if rc != 0 {
		return
	}

	fd = int(cFd)
	err = nil

	return
}

// SaveBackground saves a recording in the background.
//
// This writes an error code (or nil) to a channel upon completion.
func (context *RecordingContext) SaveBackground(filename string, complete chan<- error) {
	fd, err := context.GetSelectDescriptor()
	if err != nil {
		complete <- err
		return
	}

	err = context.SaveAsync(filename)
	if err != nil {
		complete <- err
		return
	}

	data := make([]byte, 1, 1)
	n, err := syscall.Read(fd, data)
	if err != nil {
		complete <- err
		return
	}
	if n != 1 {
		complete <- ErrSaveBackgroundReadFailed
		return
	}

	complete <- nil
}

// Discard recorded program history from memory.
//
// Recording state that is currently held in memory is freed, and may no
// longer be saved.
func (context *RecordingContext) Discard() (err error) {
	if !context.valid {
		return ErrRecordingContextDiscarded
	}
	context.valid = false

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_discard(context.ctx)
	if rc != 0 {
		return
	}
	return nil
}

// SaveOnTermination sets a recording filename to generate if  we terminate while being recorded.
//
// If the program terminates in between calls to Start and Stop
// the recorded history up to that time will be saved to a recording.
func SaveOnTermination(filename string) (err error) {
	cstring := C.CString(filename)
	defer C.free(unsafe.Pointer(cstring))

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_save_on_termination(cstring)
	if rc != 0 {
		return
	}
	return nil
}

// SaveOnTerminationCancel sancels any previous call to SaveOnTermination.
func SaveOnTerminationCancel() (err error) {
	lock.Lock()
	defer lock.Unlock()
	rc, err := C.undolr_save_on_termination_cancel()
	if rc != 0 {
		return
	}
	return nil
}

// EventLogSizeGet retrieves the current maximum size for the event log.
func EventLogSizeGet() (size int64, err error) {
	var cBytes C.long

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_event_log_size_get(&cBytes)
	if rc != 0 {
		return 0, err
	}
	return int64(cBytes), nil
}

// EventLogSizeSet set the maximum size for the event log.
func EventLogSizeSet(size int64) (err error) {
	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_event_log_size_set(C.long(size))
	if rc != 0 {
		return
	}
	return nil
}

// IncludeSymbolFiles controls whether symbol files should be included in saved recordings.
func IncludeSymbolFiles(include bool) (err error) {
	var cInclude C.int
	if include {
		cInclude = 1
	}

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_include_symbol_files(cInclude)
	if rc != 0 {
		return
	}
	return nil
}

// ShmemLogFilenameSet sets the path of the file for logging all shared memory accesses.
//
// When a shared memory log filename is set, all accesses to shared memory get logged to that
// file, which can be written by multiple processes at the same time.
// If this function is not called (or called with "" as the filename), then an external
// shared memory log is not used.
//
// This feature is currently used in the following way:
// - A process creates some shared maps.
// - It calls ShmemLogFilenameSet.
// - It forks some children processes which share the shared memory maps.
// - All the processes call Start to record themselves.
// When the processes terminate, loading one of their recordings in UndoDB will also load the
// shared memory access log. Use the <ublame> command to track cross-process accesses to an
// address in shared memory.
//
// This function must be called before recording starts or it will fail with EINVAL.
//
// Currently, recording accesses to the same map which is mapped at different addresses in
// different processes is not supported.
//
// A process is allowed to call Start and Stop multiple times and log its
// accesses to the same shared memory log. All the accesses while recording will be logged to
// the same file.
// This means that separate independent runs should not use the same shared memory log as
// the old log is not discarded for the new run.
func ShmemLogFilenameSet(filename string) (err error) {
	var cstring *C.char

	if len(filename) > 0 {
		cstring = C.CString(filename)
		defer C.free(unsafe.Pointer(cstring))
	}

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_shmem_log_filename_set(cstring)
	if rc != 0 {
		return
	}
	return nil
}

// ShmemLogFilenameClear clears the path of the file for logging shared memory accesses.
//
// This has the effect of stopping shared memory logging.
func ShmemLogFilenameClear() (err error) {
	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_shmem_log_filename_set((*C.char)(nil))
	if rc != 0 {
		return
	}
	return nil
}

// ShmemLogFilenameGet retrieves the current path for the shared memory access log.
func ShmemLogFilenameGet() (filename string, err error) {
	var cOFilename *C.char

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_shmem_log_filename_get(&cOFilename)
	if rc != 0 {
		return "", err
	}
	return C.GoString(cOFilename), nil
}

// ShmemLogSizeSet sets the maximum shared memory log access size.
func ShmemLogSizeSet(size int64) (err error) {
	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_shmem_log_size_set(C.ulong(size))
	if rc != 0 {
		return
	}
	return nil
}

// ShmemLogSizeGet retrieves the maximum shared memory log access size.
func ShmemLogSizeGet() (size int64, err error) {
	var cMaxSize C.ulong

	lock.Lock()
	defer lock.Unlock()

	rc, err := C.undolr_shmem_log_size_get(&cMaxSize)
	if rc != 0 {
		return 0, err
	}
	return int64(cMaxSize), nil
}
