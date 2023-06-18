#define DANCE_SOLVER_INTERNAL_INCLUDE
#include "dance-solver/logger.hpp"
#undef DANCE_SOLVER_INTERNAL_INCLUDE

#include <iostream>

LogStream::LogStream(logger *l, LogFunc logFunc)
    : logger_(l),
      logFunc_(logFunc) {}

LogStream::~LogStream()
{
    if (!ss_.str().empty())
    {
        logFunc_(logger_->handle, ss_.str().c_str());
    }
}
