/*
Copyright (c) 2016-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

package undoex

// -build linux

// #cgo linux,amd64 LDFLAGS: -l undoex_x64
// #cgo linux,386 LDFLAGS: -l undoex_x32
// #cgo linux,arm64 LDFLAGS: -l undoex_arm64
// #include <undoex-test-annotations.h>
// #include <stdlib.h>
// #include <errno.h>
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

// An AnnotationTestResult is used to specify the result of a test.
type AnnotationTestResult int

// Test result values for AnnotationTestResult
const (
	Unknown AnnotationTestResult = C.undoex_test_result_UNKNOWN
	Success AnnotationTestResult = C.undoex_test_result_SUCCESS
	Failure AnnotationTestResult = C.undoex_test_result_FAILURE
	Skipped AnnotationTestResult = C.undoex_test_result_SKIPPED
	Other   AnnotationTestResult = C.undoex_test_result_OTHER
)

// An AnnotationTestContext keeps track of a test run through annotations.
//
// To use, first allocate the test annotation context with
// <AnnotationTestNew>.
// To mark the start, end, result available, etc. of the test use the
// appropriate functions.
// When you are done and don't need the object any more, free with
// <Free>.
type AnnotationTestContext struct {
	ctx   *C.undoex_test_annotation_t
	valid bool
	file  string
	line  int
}

// A set of error codes returned by methods handling test annotation contexts.
var (
	ErrAnnotationTestContextInvalid = errors.New("annotation test context invalid - already freed?")
	ErrAnnotationTestResultInvalid  = errors.New("not a valid AnnotationTestResult")
	ErrAnnotationTestMissingDetail  = errors.New("detail must not be empty")
)

// AnnotationTestNew creates a context for a test that can be stored in a recording.
//
// The returned AnnotationTestContext allows details of a test run to be
// programmatically inserted in the recording.
//
// In case your program makes it possible to execute the same test twice
// during a single execution of the program, you can pass true as
// <addRunSuffix> to help disambiguate between different runs of the
// same test.
//
// The AnnotationTestContext returned must eventually be freed using Free.
func AnnotationTestNew(baseName string, addRunSuffix bool) (*AnnotationTestContext, error) {
	cName := C.CString(baseName)
	defer C.free(unsafe.Pointer(cName))

	ctx, err := C.undoex_test_annotation_new(cName, (C.bool)(addRunSuffix))
	if ctx == nil {
		return nil, err
	}

	newContext := &AnnotationTestContext{
		ctx:   ctx,
		valid: true,
	}
	_, newContext.file, newContext.line, _ = runtime.Caller(1)
	runtime.SetFinalizer(newContext, annotationTestContextFinalizer)

	return newContext, nil
}

func annotationTestContextFinalizer(context *AnnotationTestContext) {
	if context.valid {
		context.Free()
		panic(fmt.Sprintf("%s:%d: AnnotationTestContext has not been Freed",
			context.file, context.line))
	}
}

// Free an annotation as returned by <AnnotationTestNew>.
func (context *AnnotationTestContext) Free() {
	if context.valid {
		context.valid = false
		C.undoex_test_annotation_free(context.ctx)
	}
}

// Start will store an annotation for the start of the test execution.
//
// This is stored in the recording as an annotation with the test name as
// annotation name and "u-test-start" as detail. No data is associated
// with the annotation.
func (context *AnnotationTestContext) Start() error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	rc, err := C.undoex_test_annotation_start(context.ctx)
	if rc != 0 {
		return err
	}
	return nil
}

// End will store an annotation for the end of the test execution.
//
// This is stored in the recording as an annotation with the test name as
// annotation name and "u-test-end" as detail. No data is associated
// with the annotation.
//
// This function should be called as soon as the test can be considered
// terminated, even if the test result, output or other information are
// not available yet.
// It's possible to call any of the other functions operating on
// <AnnotationTestContext> after the test is marked as finished.
func (context *AnnotationTestContext) End() error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	rc, err := C.undoex_test_annotation_end(context.ctx)
	if rc != 0 {
		return err
	}
	return nil
}

// SetResult stores whether the test passed or not as an annotation in the recording.
//
// This is stored in the recording as an annotation with the test name as
// annotation name and "u-test-result" as detail. The result is stored as its
// data.
//
// You can call this function at any point after calling <Start>,
// including before or after calling <End>.
func (context *AnnotationTestContext) SetResult(result AnnotationTestResult) error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	switch result {
	case Unknown, Success, Failure, Skipped, Other:
		break
	default:
		return ErrAnnotationTestResultInvalid
	}

	rc, err := C.undoex_test_annotation_set_result(context.ctx,
		(C.undoex_test_result_t)(result))
	if rc != 0 {
		return err
	}
	return nil
}

// SetOutput stores the textual output of the test.
//
// This is stored in the recording as an annotation with the test name as
// annotation name and "u-test-output" as detail. The result is stored as
// its data.
func (context *AnnotationTestContext) SetOutput(contentType AnnotationContentType, output string) error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	switch contentType {
	case JSON, XML, UnstructuredText:
		break
	default:
		return ErrAnnotationContentTypeInvalid
	}

	cOutput := C.CString(output)
	defer C.free(unsafe.Pointer(cOutput))

	rc, err := C.undoex_test_annotation_set_output(context.ctx,
		(C.undoex_annotation_content_type_t)(contentType), cOutput)
	if rc != 0 {
		return err
	}
	return nil
}

// AddRawData adds an annotation (which stores <rawData>) at the current execution point.
//
// See <AnnotationAddRawData> for extra details.
func (context *AnnotationTestContext) AddRawData(detail string, rawData []byte) error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	if len(detail) == 0 {
		return ErrAnnotationTestMissingDetail
	}

	cDetail := C.CString(detail)
	defer C.free(unsafe.Pointer(cDetail))

	var cRawData *C.uint8_t
	cRawDataLen := (C.size_t)(0)
	if len(rawData) > 0 {
		cRawData = (*C.uint8_t)(C.CBytes(rawData))
		defer C.free(unsafe.Pointer(cRawData))
		cRawDataLen = (C.size_t)(len(rawData))
	}

	rc, err := C.undoex_test_annotation_add_raw_data(context.ctx,
		cDetail, cRawData, cRawDataLen)
	if rc != 0 {
		return err
	}
	return nil
}

// AddText adds an annotation (which stores <text>) at the current execution point.
//
// See <AnnotationAddText> for extra details.
func (context *AnnotationTestContext) AddText(detail string, contentType AnnotationContentType, text string) error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	switch contentType {
	case JSON, XML, UnstructuredText:
		break
	default:
		return ErrAnnotationContentTypeInvalid
	}

	if len(detail) == 0 {
		return ErrAnnotationTestMissingDetail
	}

	cDetail := C.CString(detail)
	defer C.free(unsafe.Pointer(cDetail))

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	rc, err := C.undoex_test_annotation_add_text(context.ctx, cDetail,
		(C.undoex_annotation_content_type_t)(contentType), cText)
	if rc != 0 {
		return err
	}
	return nil
}

// AddInt adds an annotation (which stores <value>) at the current execution point.
//
// See <AnnotationAddInt> for extra details.
func (context *AnnotationTestContext) AddInt(detail string, value int64) error {
	if !context.valid {
		return ErrAnnotationTestContextInvalid
	}

	if len(detail) == 0 {
		return ErrAnnotationTestMissingDetail
	}

	cDetail := C.CString(detail)
	defer C.free(unsafe.Pointer(cDetail))

	rc, err := C.undoex_test_annotation_add_int(context.ctx, cDetail,
		(C.int64_t)(value))
	if rc != 0 {
		return err
	}
	return nil
}
