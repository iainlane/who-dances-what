#include <stdio.h>
#include <stdlib.h>

#include "testlogger.h"

#define DEFINE_LOG_FUNC(level)                                     \
    static void log_##level(LoggerHandle handle, cchar_t *message) \
    {                                                              \
        printf(#level ": %s\n", message);                          \
    }

DEFINE_LOG_FUNC(debug);
DEFINE_LOG_FUNC(error);
DEFINE_LOG_FUNC(info);
DEFINE_LOG_FUNC(trace);
DEFINE_LOG_FUNC(warn);

logger *new_test_logger()
{
    logger *l = calloc(1, sizeof(logger));
    l->log_debug = log_debug;
    l->log_error = log_error;
    l->log_info = log_info;
    l->log_trace = log_trace;
    l->log_warn = log_warn;

    return l;
}

void free_test_logger(logger *l)
{
    free(l);
}
