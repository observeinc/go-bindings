/* Copyright (C) 2016 Undo Ltd. */

#pragma once

#include "./undoex-annotations.h"

#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

#ifndef WEAK_SYMBOL
#define WEAK_SYMBOL __attribute__((weak))
#endif

/**
 * \brief The result of a test.
 */
typedef enum
{
    /* Keep in sync with the values used in the debugger. */
    undoex_test_result_UNKNOWN, /**< The result is not known, maybe an error occurred */
    undoex_test_result_SUCCESS, /**< The test passed */
    undoex_test_result_FAILURE, /**< The test failed */
    undoex_test_result_SKIPPED, /**< The test was skipped */
    undoex_test_result_OTHER,   /**< The test result cannot be represented with this enumeration */
} undoex_test_result_t;

/**
 * \brief An object to keep track of a test run through annotations.
 *
 * To use, first allocate the test annotation object with
 * undoex_test_annotation_new(). Update the object using the functions
 * from this interface, for example undoex_test_annotation_start() to
 * set the start of the annotation. When you are done and don't need
 * the object any more, free it with undoex_test_annotation_free().
 */
typedef struct _undoex_test_annotation_t undoex_test_annotation_t;

/**
 * \brief Create an annotation for a test that can be stored in the
 * recording.
 *
 * The returned annotation allows details of a test run to be
 * programmatically inserted in the recording.
 *
 * In case your program makes it possible to execute the same test twice
 * during a single execution of the program, you can pass true as
 * `add_run_suffix` to help disambiguate between different runs of the
 * same test.
 *
 * \param base_test_name the name of the test being run
 * \param add_run_suffix if true, a suffix is added to the test_name to
 *        differentiate different runs.
 * \return a newly allocated `undoex_test_annotation_t` in case of success,
 *         NULL otherwise (and errno will be set accordingly)
 */
undoex_test_annotation_t *
WEAK_SYMBOL
undoex_test_annotation_new(const char *base_test_name,
                           bool add_run_suffix);

/**
 * \brief Free an `undoex_test_annotation_t` previously allocated with
 *        undoex_test_annotation_new().
 *
 * \param test_annotation the test annotation to free
 */
void
WEAK_SYMBOL
undoex_test_annotation_free(undoex_test_annotation_t *test_annotation);

/**
 * \brief Store an annotation for the start of the test execution.
 *
 * This is stored in the recording as an annotation with the test name as
 * annotation name and "u-test-start" as detail. No data is associated
 * with the annotation.
 *
 * \param test_annotation a test annotation
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_start(undoex_test_annotation_t *test_annotation);

/**
 * \brief Store an annotation for the end of the test execution.
 *
 * This is stored in the recording as an annotation with the test name as
 * annotation name and "u-test-end" as detail. No data is associated
 * with the annotation.
 *
 * This function should be called as soon as the test can be considered
 * terminated, even if the test result, output or other information are
 * not available yet.
 * It's possible to call any of the other functions operating on
 * `undoex_test_annotation_t` after the test is marked as finished.
 *
 * \param test_annotation a test annotation
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_end(undoex_test_annotation_t *test_annotation);

/**
 * \brief Store whether the test passed or not as an annotation in the recording.
 *
 * This is stored in the recording as an annotation with the test name as
 * annotation name and "u-test-result" as detail. The result is stored as its
 * data.
 *
 * You can call this function at any point after calling
 * undoex_test_annotation_start(), including before or after calling
 * undoex_test_annotation_end().
 *
 * \param test_annotation a test annotation
 * \param test_result a `undoex_test_result_t` indicating whether the test
 *        pass, failed, was skipped, etc.
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_set_result(undoex_test_annotation_t *test_annotation,
                                  undoex_test_result_t test_result);

/**
 * \brief Store the textual output of the test.
 *
 * This is stored in the recording as an annotation with the test name as
 * annotation name and "u-test-output" as detail. The result is stored as
 * its data.
 *
 * \param test_annotation a test annotation
 * \param content_type the type of the stored textual output
 * \param output the null-terminated string to store in the recording
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_set_output(undoex_test_annotation_t *test_annotation,
                                  undoex_annotation_content_type_t content_type,
                                  const char *output);

/**
 * \brief Add an annotation (which stores `raw_data` if not NULL) at the
 * current execution point.
 *
 * See undoex_annotation_add_raw_data() for extra details.
 *
 * \param test_annotation a test annotation
 * \param detail a string specifying extra information about the annotation;
 *        this cannot be NULL (otherwise there would be no way of
 *        distinguishing different events for this test).
 * \param raw_data the data to store in the recording or NULL to store the
 *        annotation without any associated data
 * \param raw_data_len the length (in bytes) of the `raw_data` buffer;
 *        ignored if `raw_data` is null
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_add_raw_data(undoex_test_annotation_t *test_annotation,
                                    const char *detail,
                                    const uint8_t *raw_data,
                                    size_t raw_data_len);

/**
 * \brief Add an annotation (which stores `text` if not null) at the current
 * execution point.
 *
 * See undoex_annotation_add_text() for extra details.
 *
 * \param test_annotation a test annotation
 * \param detail a string specifying extra information about the annotation;
 *        this cannot be NULL (otherwise there would be no way of
 *        distinguishing different events for this test).
 * \param content_type the type of the stored textual content
 * \param text the null-terminated string to store in the recording or NULL
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_add_text(undoex_test_annotation_t *test_annotation,
                                const char *detail,
                                undoex_annotation_content_type_t content_type,
                                const char *text);

/**
 * \brief Add an annotation (which stores `value`) at the current execution point.
 *
 * See undoex_annotation_add_int() for extra details.
 *
 * \param test_annotation a test annotation
 * \param detail a string specifying extra information about the annotation;
 *        this cannot be NULL (otherwise there would be no way of
 *        distinguishing different events for this test).
 * \param value the numeric value to store in the recording
 * \return 0 in case of success. Otherwise, -1 and `errno` will be set accordingly.
 *         In particular, if this function is called while not recording, `errno` is set to
 *         `ENOTSUP`.
 */
int
WEAK_SYMBOL
undoex_test_annotation_add_int(undoex_test_annotation_t *test_annotation,
                               const char *detail,
                               int64_t value);

#ifdef __cplusplus
}
#endif
