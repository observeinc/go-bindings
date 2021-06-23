/* Copyright (C) 2016 Undo Ltd. */

#pragma once

#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

#ifndef __WEAK_SYMBOL__
#define __WEAK_SYMBOL__ __attribute__((weak))
#endif

/**
 * Annotations are a way to programmatically insert data at the current
 * point in execution in the recording.
 * Later, the user can analyse the annotations and jump to the point in
 * execution where they were made.
 *
 * Annotations are identified by a name and an optional detail.
 *
 * Names starting with "u-" are reserved for internal use.
 *
 * The detail is useful to distinguish between different, but related,
 * annotations with the same name.
 * For instance, a test could insert an annotation (with its name as name)
 * when it starts and another one when it ends. The details can mark the
 * beginning of the test and the end of the test.
 *
 * Note that a NULL detail is considered equivalent to a detail identified
 * by the empty string.
 *
 * Each annotation may be associated with some content. The content can be
 * arbitrary binary data or other predefined types. If a predefined type
 * matches your content type, it's better to use the predefined type instead of
 * just storing the content as binary data. For instance, if you are storing
 * some JSON, use undoex_annotation_add_text() with
 * `undoex_annotation_content_type_JSON`, if you need to store an int use
 * undoex_annotation_add_int(). This allows the debugger to present the data
 * more appropriately.
 */

/**
 * \brief The type of text data stored in a recording.
 *
 * See undoex_annotation_add_text() for details on the usage.
 */
typedef enum
{
    undoex_annotation_content_type_JSON              = 101, /**< JSON text */
    undoex_annotation_content_type_XML               = 102, /**< XML text */
    undoex_annotation_content_type_UNSTRUCTURED_TEXT = 100, /**< Plain text not matching any other format */
} undoex_annotation_content_type_t;

/**
 * \brief Add an annotation (which stores `raw_data` if not NULL) at the
 * current execution point.
 *
 * The stored data can contain any sequence of bytes (including the zero byte).
 *
 * If your data is textual add undoex_annotation_add_text() instead. If it's
 * numeric use undoex_annotation_add_int().
 *
 * \param name the name of the annotation
 * \param detail an optional string specifying extra information about this
 *        specific annotation (or NULL)
 * \param raw_data the data to store in the recording or NULL to store the
 *        annotation without any associated data
 * \param raw_data_len the length (in bytes) of the `raw_data` buffer;
 *        ignored if `raw_data` is null
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
__WEAK_SYMBOL__
undoex_annotation_add_raw_data(const char *name,
                               const char *detail,
                               const uint8_t *raw_data,
                               size_t raw_data_len);

/**
 * \brief Add an annotation (which stores `text` if not null) at the current
 * execution point.
 *
 * The stored data is a string terminated by a zero byte. If you need to store
 * arbitrary data including null characters, use
 * undoex_annotation_add_raw_data() instead.
 *
 * By specifying the type of the textual content, you allow the debugger to
 * display the content in a smarter way.
 *
 * \param name the name of the annotation
 * \param detail an optional string specifying extra information about this
 *        specific annotation (or NULL)
 * \param content_type the type of the stored textual content
 * \param text the null-terminated string to store in the recording or NULL
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
__WEAK_SYMBOL__
undoex_annotation_add_text(const char *name,
                           const char *detail,
                           undoex_annotation_content_type_t content_type,
                           const char *text);

/**
 * \brief Add an annotation (which stores `value`) at the current execution point.
 *
 * \param name the name of the annotation
 * \param detail an optional string specifying extra information about this
 *        specific annotation (or NULL)
 * \param value the numeric value to store in the recording
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
__WEAK_SYMBOL__
undoex_annotation_add_int(const char *name,
                          const char *detail,
                          int64_t value);

#ifdef __cplusplus
}
#endif
