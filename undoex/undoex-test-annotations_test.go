/*
Copyright (c) 2016-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

package undoex

import (
	"runtime"
	"testing"
	"time"
)

func TestAnnotationTestNew(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	context.Free()
}

func TestAnnotationTestStartEnd(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Free()

	err = context.Start()
	if err != nil {
		t.Fatal(err)
	}

	err = context.End()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationTestSetResult(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Free()

	err = context.SetResult(Unknown)
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetResult(Success)
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetResult(Failure)
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetResult(Skipped)
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetResult(Other)
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetResult(42)
	if err != ErrAnnotationTestResultInvalid {
		t.Fatal("Expected SetResult() to fail with invalid result")
	}

}

func TestAnnotationTestSetOutput(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Free()

	err = context.SetOutput(JSON, "")
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetOutput(XML, "")
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetOutput(UnstructuredText, "")
	if err != nil {
		t.Fatal(err)
	}

	err = context.SetOutput(42, "")
	if err != ErrAnnotationContentTypeInvalid {
		t.Fatal("Expected SetOutput() to fail with invalid content type")
	}
}

func TestAnnotationTestAdd(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Free()

	err = context.AddRawData("d", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = context.AddRawData("d", []byte{42, 42, 42, 42})
	if err != nil {
		t.Fatal(err)
	}

	err = context.AddText("d", JSON, "")
	if err != nil {
		t.Fatal(err)
	}

	err = context.AddText("d", XML, "")
	if err != nil {
		t.Fatal(err)
	}

	err = context.AddText("d", UnstructuredText, "")
	if err != nil {
		t.Fatal(err)
	}

	err = context.AddText("d", 42, "")
	if err != ErrAnnotationContentTypeInvalid {
		t.Fatal("Expected AddText() to fail with invalid content type")
	}

	err = context.AddInt("d", 42)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAnnotationTestUseAfterFree(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	context.Free()

	err = context.Start()
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected Start() to fail with use after free")
	}

	err = context.End()
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected End() to fail with use after free")
	}

	err = context.SetResult(Unknown)
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected SetResult() to fail with use after free")
	}

	err = context.SetOutput(XML, "")
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected SetOutput() to fail with use after free")
	}

	err = context.AddRawData("d", nil)
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected AddRawData() to fail with use after free")
	}

	err = context.AddText("d", XML, "")
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected AddText() to fail with use after free")
	}

	err = context.AddInt("d", 42)
	if err != ErrAnnotationTestContextInvalid {
		t.Fatal("Expected AddInt() to fail with use after free")
	}
}

func TestAnnotationTestMissingDetail(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}
	defer context.Free()

	err = context.AddRawData("", nil)
	if err != ErrAnnotationTestMissingDetail {
		t.Fatal("Expected AddRawData() to fail with missing detail")
	}

	err = context.AddText("", XML, "")
	if err != ErrAnnotationTestMissingDetail {
		t.Fatal("Expected AddText() to fail with missing detail")
	}

	err = context.AddInt("", 42)
	if err != ErrAnnotationTestMissingDetail {
		t.Fatal("Expected AddInt() to fail with missing detail")
	}
}

func TestAnnotationTestLeak(t *testing.T) {
	context, err := AnnotationTestNew("testname", false)
	if err != nil {
		t.Fatal(err)
	}

	// Wrap the finalizer for the context with a function:
	//  - checks the finalizer panics
	//  - sends to a channel
	ch := make(chan bool, 1)

	runtime.SetFinalizer(context, nil)
	runtime.SetFinalizer(context, func(c *AnnotationTestContext) {
		defer func() {
			ch <- true
			if r := recover(); r == nil {
				t.Fatalf("Finalizer didn't panic")
			}
		}()
		annotationTestContextFinalizer(c)
	})

	// This should trigger the Finalizer, and we didn't call Free
	context = nil
	runtime.GC()

	select {
	case <-ch:
	case <-time.After(time.Second * 10):
		t.Fatalf("Finalizer hadn't run after 10 seconds")
	}
}
