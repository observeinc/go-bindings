/** @file */

#pragma once


#ifdef __cplusplus
extern "C" {
#endif

#ifndef USE_WEAK_SYMBOLS
#define __WEAK_SYMBOL__
#else
#define __WEAK_SYMBOL__ __attribute__((weak))
#endif

#ifdef __cplusplus
}
#endif
