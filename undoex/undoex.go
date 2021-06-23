/*
Copyright (c) 2016-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

package undoex

// #include <undoex-annotations.h>
// #include <stdlib.h>
// #include <errno.h>
import "C"
import (
	"errors"
	"unsafe"
)

// An AnnotationContentType identifies the type of textual context to be stored in a recording.
type AnnotationContentType int

// Content type values for AnnotationContentType
const (
	JSON             AnnotationContentType = C.undoex_annotation_content_type_JSON
	XML              AnnotationContentType = C.undoex_annotation_content_type_XML
	UnstructuredText AnnotationContentType = C.undoex_annotation_content_type_UNSTRUCTURED_TEXT
)

// ErrAnnotationContentTypeInvalid indicates the content type is outside the valid range.
var ErrAnnotationContentTypeInvalid = errors.New("content type not valid")

// AnnotationAddRawData adds an annotation (which stores <raw_data> if not NULL) at the current execution point.
//
// The stored data can contain any sequence of bytes (including '\0').
//
// If your data is textual add AnnotationAddText() instead. If it's
// numeric use AnnotationAddInt().
func AnnotationAddRawData(name, detail string, rawData []byte) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cDetail *C.char
	if len(detail) > 0 {
		cDetail = C.CString(detail)
		defer C.free(unsafe.Pointer(cDetail))
	}

	cRawData := (*C.uint8_t)(C.CBytes(rawData))
	defer C.free(unsafe.Pointer(cRawData))

	cRawDataLen := (C.size_t)(len(rawData))

	rc, err := C.undoex_annotation_add_raw_data(cName, cDetail, cRawData, cRawDataLen)
	if rc != 0 {
		return err
	}
	return nil
}

// AnnotationAddText adds an annotation (which stores <text> if not null) at the current execution point.
//
// The stored data is a string terminated by a '\0'. If you need to store
// data including null characters, use <AnnotationAddRawData> instead.
//
// By specifying the type of the textual content, you allow the debugger to
// display the content in a smarter way.
func AnnotationAddText(name, detail string, contentType AnnotationContentType, text string) error {
	switch contentType {
	case JSON, XML, UnstructuredText:
		break
	default:
		return ErrAnnotationContentTypeInvalid
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cDetail *C.char
	if len(detail) > 0 {
		cDetail = C.CString(detail)
		defer C.free(unsafe.Pointer(cDetail))
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	rc, err := C.undoex_annotation_add_text(cName, cDetail,
		(C.undoex_annotation_content_type_t)(contentType), cText)
	if rc != 0 {
		return err
	}
	return nil
}

// AnnotationAddInt adds an annotation (which stores <value>) at the current execution point.
func AnnotationAddInt(name, detail string, value int64) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cDetail *C.char
	if len(detail) > 0 {
		cDetail = C.CString(detail)
		defer C.free(unsafe.Pointer(cDetail))
	}

	rc, err := C.undoex_annotation_add_int(cName, cDetail,
		(C.int64_t)(value))
	if rc != 0 {
		return err
	}
	return nil
}
