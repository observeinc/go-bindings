/** @file */

/* Copyright (C) 2020 Undo Ltd. */

#ifndef UNDOLR_H
#define UNDOLR_H

#include "../common/common.h"

#ifdef __cplusplus
extern "C" {
#endif

/**
 * API for the Undo Live Recorder.
 *
 * Allows an application to create a LiveRecorder recording of itself running, which can
 * then be loaded in UDB.
 *
 * Except where documented, functions in this API return 0 in case of success,
 * or -1 in case of failure, with an appropriate error code in \c errno.
 *
 * Calling undolr_*() functions concurrently from different threads is not
 * supported, and will give undefined behaviour.
 *
 * Please also note that if the application calls undolr_*() functions in one
 * thread while another thread is waiting for child process by calling
 * waitpid(-1,&status,__WALL), the application may hang.  The issue arises because
 * Live Recorder needs to create and wait for a short-lived child process. This
 * child process is created as a 'clone' process so that it will only be visible
 * if the __WALL flag is specified.
 */

/**
 * \brief Recording context.
 *
 * This type serves as a handle for a recorded session whilst it is still in
 * memory, valid between matched calls to undolr_stop() and undolr_discard().
 */
typedef struct undolr_recording_context_private_t *undolr_recording_context_t;

/**
 * \brief Reason for failing to start recording.
 *
 * See undolr_start() and undolr_error_string().
 */
typedef enum
{
    undolr_error_NONE = 0,

    undolr_error_NO_ATTACH_YAMA = 1,
    /**< Failure to start-up due to failure to attach to the application process
    due to /proc/sys/kernel/yama/ptrace_scope. */

    undolr_error_CANNOT_ATTACH = 2,

    undolr_error_LIBRARY_SEARCH_FAILED = 3,
    /**< Failed to find dynamic libraries used by application. */

    undolr_error_CANNOT_RECORD = 4,
    /**< Miscellaneous errors without specific error codes. */

    undolr_error_NO_THREAD_INFO = 5,
    /**< Live Recorder was unable to find information about threads. */

    undolr_error_PKEYS_IN_USE = 6
    /**< Use of Protection Keys was detected. This is not yet supported. */
} undolr_error_t;

/**
 * \brief Return string describing error number.
 *
 * \param[in] error Error number from a call to undolr_start().
 *
 * \return String describing the error, or "<unknown error>".
 */
const char * __WEAK_SYMBOL__ undolr_error_string(undolr_error_t error);

/**
 * \brief Start recording the current process.
 *
 * The current process must not already be being recorded: that is, either
 * undolr_start() is being called for the first time, or else there was a call
 * to undolr_stop() since the most recent call to undolr_start().
 *
 * \param error [out] pointer to a location to store the reason for
 *        failure to start recording, or \c NULL not to receive this.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 *
 * The caller must check the return value to determine if the call succeeded or
 * failed. If it failed, there may be more specific information about the
 * failure in \c *error.
 */
int __WEAK_SYMBOL__ undolr_start(undolr_error_t *error);

/**
 * \brief Get the version string for this release.
 */
const char * __WEAK_SYMBOL__ undolr_get_version_string(void);

/**
 * \brief Stop recording the current process.
 *
 * \param context [in] pointer to a location to store the recording context, or
 *        \c NULL if the recording context should be immediately discarded.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 *
 * If not discarded immediately, the recorded history is held in memory until
 * \c context is passed to undolr_discard().
 */
int __WEAK_SYMBOL__ undolr_stop(undolr_recording_context_t *context);

/**
 * \brief Save recorded program history to a named recording file.
 *
 * The current process must be being recorded: that is, undolr_start() must
 * have been successfully invoked without a subsequent call to undolr_stop().
 *
 * \param filename [in] name of the recording file.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 *
 * undolr_save() may be called any number of times until undolr_stop() is
 * called. Each subsequent call to undolr_save() contains later execution
 * history. The recordings are independent of each other, and each can be
 * replayed on its own.
 */
int __WEAK_SYMBOL__ undolr_save(const char *filename);

/**
 * \brief Start asynchronously saving recorded program history to a named
 *        recording file.
 *
 * \param context [in] recording context returned by a previous call to
 *        undolr_stop().
 *
 * \param filename [in] name of the recording file.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 *
 * After this call, the recording context may be passed to
 * undolr_poll_saving_progress(), but must not be passed to undolr_discard() or
 * to another call to undolr_save_async() until the save has completed.
 */
int __WEAK_SYMBOL__ undolr_save_async(undolr_recording_context_t context, const char *filename);

/**
 * \brief Check the status of an asynchronous save operation.
 *
 * \param context [in] recording context passed to undolr_save_async().
 *
 * \param complete [out] pointer to a location that is updated to zero if the
 *        save is still in progress, or non-zero if it is complete. It is not
 *        updated if the status could not be determined.
 *
 * \param result [out] pointer to a location that is updated only if the save
 *        operation is complete. It is set to 0 if the recording was saved
 *        successfully, or to an appropriate error code if it was not. It is
 *        not updated if \c *complete is zero or if the status could not be
 *        determined.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_poll_saving_complete(undolr_recording_context_t context,
                                int *complete, int *result);

/**
 * \brief Check the status and progress of an asynchronous save operation.
 *
 * \param context [in] recording context passed to undolr_save_async().
 *
 * \param complete [out] pointer to a location that is updated to zero if the
 *        save is still in progress, or non-zero if it is complete. It is not
 *        updated if the status could not be determined.
 *
 * \param progress [out] pointer to a location that is updated only if the save
 *        is still in progress. It is set to the percentage of completion,
 *        rounded down, from 0 to 100 inclusive, or to -1 if progress
 *        information is unavailable. It is not updated if \c *complete is
 *        non-zero or if the status could not be determined.
 *
 * \param result [out] pointer to a location that is updated only if the save
 *        operation is complete. It is set to 0 if the recording was saved
 *        successfully, or to an appropriate error code if it was not. It is
 *        not updated if \c *complete is zero or if the status could not be
 *        determined.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_poll_saving_progress(undolr_recording_context_t context,
                                int *complete, int *progress, int *result);

/**
 * \brief Get a selectable file descriptor to detect save state changes.
 *
 * \param context [in] recording context passed to undolr_save_async().
 *
 * \param fd [out] pointer to a location to store a file descriptor.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 *
 * When the asynchronous save operation is complete, a byte is written to the
 * file descriptor, allowing it to be selected for read using select() or
 * pselect().
 *
 * The file descriptor is closed when \c context is passed to undolr_discard().
 */
int __WEAK_SYMBOL__ undolr_get_select_descriptor(undolr_recording_context_t context, int *fd);

/**
 * \brief Discard recorded program history from memory.
 *
 * \param context [in] recording context returned by a previous call to
 *        undolr_stop().
 *
 * After calling this, the memory used to maintain the recording state has been
 * freed, and \c context must not be passed to any other function in this API.
 */
int __WEAK_SYMBOL__ undolr_discard(undolr_recording_context_t context);

/**
 * Instruct LiveRecorder to save a recording when the current process exits.
 *
 * LiveRecorder must have been started by a successful call to undolr_start()
 * before calling this function.
 *
 * \param filename [in] the name of the recording file.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 *
 * The instruction is cancelled by a call to
 * undolr_save_on_termination_cancel() or undolr_stop().
 */
int __WEAK_SYMBOL__ undolr_save_on_termination(const char *filename);

/**
 * \brief Cancel any previous call to undolr_save_on_termination().
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_save_on_termination_cancel(void);

/**
 * \brief Retrieve the current event log size
 *
 * \param bytes current event log size is stored here on success.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_event_log_size_get(long *bytes);

/**
 * \brief Set the event log size.
 *
 * \param bytes new event log size.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_event_log_size_set(long bytes);

/**
 * \brief Control whether to include symbol files in saved recordings.
 *
 * \param include non-zero to include symbol files, zero to skip. Default 1.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_include_symbol_files(int include);

/**
 * \brief Set the path of the file where to log all shared memory accesses.
 *
 * When a shared memory log filename is set, all accesses to shared memory get logged to that
 * file, which can be written by multiple processes at the same time.
 * If this function is not called (or called with `NULL` as `filename`), then an external
 * shared memory log is not used.
 *
 * This feature is currently used in the following way:
 * - A process creates some shared maps.
 * - It calls undolr_shmem_log_filename_set().
 * - It forks some children processes which share the shared memory maps.
 * - All the processes call undolr_start() to record themselves.
 * When the processes terminate, loading one of their recordings in UDB will also load the
 * shared memory access log. Use the `ublame` command to track cross-process accesses to an
 * address in shared memory.
 *
 * This function must be called before recording starts or it will fail and set `errno` to
 * `EINVAL`.
 *
 * Currently, recording accesses to the same map which is mapped at different addresses in
 * different processes is not supported.
 *
 * A process is allowed to call undolr_start() and undolr_stop() multiple times and log its
 * accesses to the same shared memory log. All the accesses while recording will be logged to
 * the same file.
 * This means that separate independent runs should not use the same shared memory log as
 * the old log is not discarded for the new run.
 *
 * \param   filename    The path to a file to use to log accesses to shared memory, or `NULL`
 *                      to disable the feature (the default).
 *                      If a non-null path is used, it must have a `.shmem` extension to
 *                      allow UDB to later find the file. If it doesn't, this function will
 *                      fail and set `errno` to `EINVAL`.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_shmem_log_filename_set(const char *filename);

/**
 * \brief Get the current path for the shared memory access log.
 *
 * See undolr_shmem_log_filename_set() for details.
 *
 * \param[out]  o_filename  A pointer to set to the currently set shared memory log, which can
 *                          be `NULL` if this feature is not enabled.
 *                          The string is valid until the next call to
 *                          undolr_shmem_log_filename_set().
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_shmem_log_filename_get(const char **o_filename);

/**
 * \brief Set the maximum size of the the file where shared memory accesses are logged.
 *
 * See undolr_shmem_log_filename_get() for details about the shared memory log.
 *
 * If this function is not called or called with `0` as argument, then a suitable default
 * value will be used.
 *
 * This function must be called before recording starts or it will fail and set `errno` to
 * `EINVAL`.
 *
 * \param   max_size    The maximum size (in bytes) for the shared memory log file.
 *                      If this is not a multiple of the page size, the actual size can be
 *                      rounded up to the next multiple.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_shmem_log_size_set(unsigned long max_size);

/**
 * \brief Get the current maximum size for the shared memory access log.
 *
 * See undolr_shmem_log_size_set() for details.
 *
 * \param[out]  o_max_size  A pointer to set to the currently set maximum shared memory log
 *                          size.
 *
 * \return 0 for success, or -1 for failure, with an error code in \c errno.
 */
int __WEAK_SYMBOL__ undolr_shmem_log_size_get(unsigned long *o_max_size);

#include "./undolr_deprecated.h"

#ifdef __cplusplus
}
#endif

#endif
