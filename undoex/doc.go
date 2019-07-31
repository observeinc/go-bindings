/*
Copyright (c) 2016-2019, Undo Ltd.
All rights reserved.

SPDX-License-Identifier: BSD-3-Clause
*/

// Package undoex provides a way to insert annotations in an UndoDB recording.
//
// Annotations are a way to programmatically insert data at the current
// point in execution in the recording. Later, the user can analyse the
// annotations and jump to the point in execution were they were made.
//
// Annotations are identified by a name and an optional detail.
//
// Names starting with "u-" are reserved for internal use.
//
// The detail is useful to distinguish between different, but related,
// annotations with the same name.
// For instance, a test could insert an annotation (with its name as name)
// when it starts and another one when it ends. The details can mark the
// beginning of the test and the end of the test.
//
// Each annotation may be associated with some content.
// The content can be arbitrary binary data or other predefined types.
// If a predefined type matches your content type, it's better to use the
// predefined type instead of just storing the content as binary data.
// For instance, if you are storing some JSON, use
// <AnnotationAddText> with <AnnotationContentType>,
// if you need to store an int use <AnnotationAddInt>.
// This allows the debugger to present the data more appropriately.
//
// The AnnotationTest* API provides a wrapper for the basic annotation
// interface with helpers for recording test running and results.
//
package undoex
