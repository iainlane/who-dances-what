package loggerbinding

/*
#cgo CFLAGS: -I${SRCDIR}/../../cppsolver/src/include -I${SRCDIR}/../../cppsolver/build/

#include <stdint.h>
#include <stdlib.h>

#define DANCE_SOLVER_INTERNAL_INCLUDE
#include "dance-solver/logger.h"
#undef DANCE_SOLVER_INTERNAL_INCLUDE

#ifdef __cplusplus
extern "C"
{
#endif

#ifndef WHO_DANCES_WHAT_LOGGER_BINDING_H
#define WHO_DANCES_WHAT_LOGGER_BINDING_H

extern void LogTrace(LoggerHandle lh, cchar_t *msg);
extern void LogDebug(LoggerHandle lh, cchar_t *msg);
extern void LogInfo(LoggerHandle lh, cchar_t *msg);
extern void LogWarn(LoggerHandle lh, cchar_t *msg);
extern void LogError(LoggerHandle lh, cchar_t *msg);

#endif //WHO_DANCES_WHAT_LOGGER_BINDING_H

#ifdef __cplusplus
} // extern "C"
#endif
*/
import "C"
import (
	"runtime/cgo"
	"unsafe"

	"github.com/sirupsen/logrus"
)

func log(lh C.LoggerHandle, level logrus.Level, msg *C.cchar_t) {
	g := (*C.char)(unsafe.Pointer(msg))
	str := C.GoString(g)

	handle := cgo.Handle(lh)
	entry, ok := handle.Value().(*logrus.Entry)
	if !ok {
		logrus.Fatalf("handle %v is not a *logrus.Entry", handle)
	}

	entry.Log(level, str)
}

//export LogTrace
func LogTrace(lh C.LoggerHandle, msg *C.cchar_t) {
	log(lh, logrus.TraceLevel, msg)
}

//export LogDebug
func LogDebug(lh C.LoggerHandle, msg *C.cchar_t) {
	log(lh, logrus.DebugLevel, msg)
}

//export LogInfo
func LogInfo(lh C.LoggerHandle, msg *C.cchar_t) {
	log(lh, logrus.InfoLevel, msg)
}

//export LogWarn
func LogWarn(lh C.LoggerHandle, msg *C.cchar_t) {
	log(lh, logrus.WarnLevel, msg)
}

//export LogError
func LogError(lh C.LoggerHandle, msg *C.cchar_t) {
	log(lh, logrus.ErrorLevel, msg)
}

func PopulateLogger(loggerHandle cgo.Handle) *C.logger {
	logger := (*C.logger)(C.malloc(C.sizeof_logger))
	logger.handle = C.uintptr_t(loggerHandle)
	logger.log_trace = (C.LogFunc)(C.LogTrace)
	logger.log_debug = (C.LogFunc)(C.LogDebug)
	logger.log_info = (C.LogFunc)(C.LogInfo)
	logger.log_warn = (C.LogFunc)(C.LogWarn)
	logger.log_error = (C.LogFunc)(C.LogError)

	return logger
}
