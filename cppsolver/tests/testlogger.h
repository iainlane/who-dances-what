#define DANCE_SOLVER_INTERNAL_INCLUDE
#include "dance-solver/logger.h"
#undef DANCE_SOLVER_INTERNAL_INCLUDE

#ifdef __cplusplus
extern "C"
{
#endif // __cplusplus
    logger *new_test_logger();
    void free_test_logger(logger *l);
#ifdef __cplusplus
}
#endif // __cplusplus
