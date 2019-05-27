/*
 * contracts.h
 *
 * Use contracts for debugging to ensure that pre- and post-conditions are met
 * in each function. Based on a similar file written by Frank Pfenning for
 * 15-122 at Carnegie Mellon University.
 *
 * To enable contracts, use the -DDEBUG flag when compiling, or run make debug
 *
 * Created by Jacob Strieb
 */

#include <assert.h>

#ifdef DEBUG

#define DEBUG 1

#define REQUIRES assert
#define ENSURES assert
#define ASSERT assert

#else /* DEBUG */

#define DEBUG 0

#define REQUIRES (void)
#define ENSURES (void)
#define ASSERT (void)

#endif /* DEBUG */
