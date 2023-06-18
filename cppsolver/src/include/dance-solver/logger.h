#pragma once

#include <stdint.h>

// don't allow this to be included directly, users must include either
// dance_solver.h or dance_solver.hpp
#ifndef DANCE_SOLVER_INTERNAL_INCLUDE
#error "Please include dance_solver.h or dance_solver.hpp instead of __FILE__"
#endif

typedef const char cchar_t;
typedef uintptr_t LoggerHandle;
typedef void (*LogFunc)(LoggerHandle, cchar_t *);

#ifdef __cplusplus
extern "C"
{
#endif // __cplusplus
    typedef struct
    {
        LoggerHandle handle;

        LogFunc log_trace;
        LogFunc log_debug;
        LogFunc log_info;
        LogFunc log_warn;
        LogFunc log_error;
    } logger;
#ifdef __cplusplus
}
#endif // __cplusplus
