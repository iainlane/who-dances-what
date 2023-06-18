#pragma once

#ifndef DANCE_SOLVER_INTERNAL_INCLUDE
#error "Please include dance_solver.h or dance_solver.hpp instead of logstream.h"
#endif

#include <memory>
#include <sstream>
#include <string>

#include "logger.h"

#define Trace(logger) LogStream(logger, logger->log_trace)
#define Debug(logger) LogStream(logger, logger->log_debug)
#define Warn(logger) LogStream(logger, logger->log_warn)
#define Info(logger) LogStream(logger, logger->log_info)
#define Error(logger) LogStream(logger, logger->log_error)

class LogStream
{
public:
    LogStream(logger *l, LogFunc logFunc);
    ~LogStream();

    template <typename T>
    LogStream &operator<<(const T &data)
    {
        ss_ << data;
        return *this;
    }

protected:
    logger *logger_;
    std::stringstream ss_;
    LogFunc logFunc_;
};
