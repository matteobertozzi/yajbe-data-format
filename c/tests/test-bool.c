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

#include <stdio.h>

#include "yajbe.h"
#include "utest.h"

static int test_yajbe_bool_simple (z_utest_t *utest) {
    z_mem_bytes_writer_t writer;
    yajbe_encoder_t encoder;
    uint8_t buffer[32];

    z_mem_bytes_writer_init_fixed(&writer, buffer, sizeof(buffer));
    yajbe_encoder_init(&encoder, NULL, &writer.super);

    yajbe_encode_true(&encoder);
    z_assert_eq(utest, 1, writer.offset);
    z_assert_eq(utest, 0x03, buffer[0]);

    yajbe_encode_false(&encoder);
    z_assert_eq(utest, 2, writer.offset);
    z_assert_eq(utest, 0x03, buffer[0]);
    z_assert_eq(utest, 0x02, buffer[1]);

    yajbe_encode_bool(&encoder, 1);
    z_assert_eq(utest, 3, writer.offset);
    z_assert_eq(utest, 0x03, buffer[0]);
    z_assert_eq(utest, 0x02, buffer[1]);
    z_assert_eq(utest, 0x03, buffer[2]);

    yajbe_encode_bool(&encoder, 0);
    z_assert_eq(utest, 4, writer.offset);
    z_assert_eq(utest, 0x03, buffer[0]);
    z_assert_eq(utest, 0x02, buffer[1]);
    z_assert_eq(utest, 0x03, buffer[2]);
    z_assert_eq(utest, 0x02, buffer[3]);

    //dump_hex(buffer, writer.offset);
    return 0;
}

int main (int argc, char **argv) {
    z_utest_t utest;

    z_utest_init(&utest);
    test_yajbe_bool_simple(&utest);
    z_utest_dump_result(&utest);
    return 0;
}
