/*
Copyright (c) 2016-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

package undoex

import (
	"testing"
)

func TestAnnotationAddRawData(t *testing.T) {
	data := []byte{42, 42, 42, 42}
	err := AnnotationAddRawData(
		"testname", "testdetail", data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationAddTextJson(t *testing.T) {
	text := "{\"test\": {\"key1\": \"value1\"}}"
	err := AnnotationAddText(
		"testname", "testdetail", JSON, text)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationAddTextXml(t *testing.T) {
	text := "<test><item id='key1'>value1</item></test>"
	err := AnnotationAddText(
		"testname", "testdetail", XML, text)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationAddTextUnstructured(t *testing.T) {
	text := "key1: value1"
	err := AnnotationAddText(
		"testname", "testdetail", UnstructuredText, text)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationAddTextInvalid(t *testing.T) {
	text := "junk"
	err := AnnotationAddText(
		"testname", "testdetail", 42, text)
	if err == nil {
		t.Fatal("Unexpected success with invalid content type")
	} else if err != ErrAnnotationContentTypeInvalid {
		t.Fatal(err)
	}
}

func TestAnnotationAddInt(t *testing.T) {
	err := AnnotationAddInt(
		"testname", "testdetail", 42)
	if err != nil {
		t.Fatal(err)
	}
}
