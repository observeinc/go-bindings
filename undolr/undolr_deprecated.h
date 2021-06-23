/** @file */

/* Copyright (C) 2017-2018 Undo Ltd. */

#ifndef UNDOLR_DEPRECATED_H
#define UNDOLR_DEPRECATED_H

#include "../common/common.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * Deprecated components of the Undo Live Recorder API.
 *
 * Unless otherwise stated, it is intended that there is an equivalent way
 * to achieve all functionality offered by functions within this header file.
 * As such, these functions are intended to be removed in later releases of
 * Undo Live Recorder.
 *
 * If you rely on behaviour that cannot obviously be fulfilled by the contents
 * of undolr.h, please contact support@undo.io for clarification.
 */

/**
 * \deprecated Deprecated alternative to undolr_start().
 *
 * Behaves identically to undolr_start() except that if recording is already in
 * operation, undolr_recording_start() returns zero (indicating success),
 * whereas undolr_start() returns -1 (indicating failure).
 */
int __WEAK_SYMBOL__ undolr_recording_start(undolr_error_t* o_error);

/**
 * \deprecated Deprecated alternative to undolr_stop().
 *
 * Behaves identically to `undolr_stop(NULL)`, so that it stops recording and
 * discards the recording context without saving.
 */
int __WEAK_SYMBOL__ undolr_recording_stop(void);

/**
 * \deprecated Deprecated alternative to undolr_stop() followed by
 * undolr_save_async().
 *
 * Stops recording, saves asynchronously to `filename`, and detaches from the
 * debuggee so that recording cannot be restarted.
 */
int __WEAK_SYMBOL__ undolr_recording_stop_and_save(const char* filename);

#ifdef __cplusplus
}
#endif

#endif
