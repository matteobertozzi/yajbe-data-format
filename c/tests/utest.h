/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef __Z_UTEST_H__
#define __Z_UTEST_H__

#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <math.h>

typedef struct {
    const char *failed_func;
    int failed_line;
    size_t succeded;
} z_utest_t;

#define z_utest_init(self)                              \
    do {                                                \
        (self)->failed_func = NULL;                     \
        (self)->failed_line = 0;                        \
        (self)->succeded = 0;                           \
    } while (0)

#define z_utest_mark_failure(self)                      \
    do {                                                \
        (self)->failed_func = __FUNCTION__;             \
        (self)->failed_line = __LINE__;                 \
    } while (0)

#define z_utest_dump_result(self)                                   \
    do {                                                            \
        if ((self)->failed_line) {                                  \
            printf(" [!!] succeded: %zu, failed %s():%d\n",         \
                (self)->succeded,                                   \
                (self)->failed_func,                                \
                (self)->failed_line);                               \
        } else {                                                    \
            printf(" [OK] succeded: %zu\n",                         \
                (self)->succeded);                                  \
        }                                                           \
    } while (0)

#define z_assert_eq(self, expected, current)            \
    do {                                                \
        if ((expected) != (current)) {                  \
            z_utest_mark_failure(self);                 \
            return -1;                                  \
        }                                               \
    } while (0);

#define z_assert_eq_float(self, expected, current, epsilon)     \
    do {                                                        \
        if (fabs((expected) - (current)) >= epsilon) {          \
            z_utest_mark_failure(self);                         \
            return -1;                                          \
        }                                                       \
    } while (0)

#define z_assert_eq_data(self, a, _alen, b, _blen)      \
    do {                                                \
        const size_t alen = _alen;                      \
        const size_t blen = _blen;                      \
        if (alen != blen) {                             \
            z_utest_mark_failure(self);                 \
            return -1;                                  \
        }                                               \
                                                        \
        if (memcmp(a, b, alen)) {                       \
            z_utest_mark_failure(self);                 \
            return -1;                                  \
        }                                               \
        (self)->succeded++;                             \
    } while (0);

#endif /* !__Z_UTEST_H__ */
