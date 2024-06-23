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

struct test_int_value {
    int64_t value;
    uint8_t buffer[5];
    int buflen;
};

static const struct test_int_value test_values[] = {
    { .value = 0, .buflen = 1, .buffer = { 0x60 } },
    { .value = 1, .buflen = 1, .buffer = { 0x40 } },
    { .value = 7, .buflen = 1, .buffer = { 0x46 } },
    { .value = 24, .buflen = 1, .buffer = { 0x57 }},
    { .value = 25, .buflen = 2, .buffer = { 0x58, 0x00 }},
    { .value = 0xff, .buflen = 2, .buffer = { 0x58, 0xe6 }},
    { .value = 0xffff, .buflen = 3, .buffer = { 0x59, 0xe6, 0xff }},
    { .value = 0xffffff, .buflen = 4, .buffer = { 0x5a, 0xe6, 0xff, 0xff }},

    { .value = -1, .buflen = 1, .buffer = { 0x61 } },
    { .value = -7, .buflen = 1, .buffer = { 0x67 } },
    { .value = -23, .buflen = 1, .buffer = { 0x77 } },
    { .value = -24, .buflen = 2, .buffer = { 0x78, 0x00 } },
    { .value = -25, .buflen = 2, .buffer = { 0x78, 0x01 } },
    { .value = -0xff, .buflen = 2, .buffer = { 0x78, 0xe7 } },
    { .value = -0xffff, .buflen = 3, .buffer = { 0x79, 0xe7, 0xff } },
};

static int test_yajbe_int_simple (z_utest_t *utest) {
    z_mem_bytes_writer_t writer;
    z_mem_bytes_reader_t reader;
    yajbe_encoder_t encoder;
    yajbe_decoder_t decoder;
    uint8_t buffer[32];

    z_mem_bytes_writer_init_fixed(&writer, buffer, sizeof(buffer));
    yajbe_encoder_init(&encoder, NULL, &writer.super);

    for (int i = 0, n = sizeof(test_values) / sizeof(struct test_int_value); i < n; ++i) {
        const struct test_int_value *ptest = test_values + i;

        // encode
        writer.offset = 0;
        yajbe_encode_int(&encoder, ptest->value);
        z_assert_eq(utest, ptest->buflen, writer.offset);
        z_assert_eq_data(utest, ptest->buffer, ptest->buflen, buffer, writer.offset);

        // decode
        z_mem_bytes_reader_init(&reader, buffer, writer.offset);
        yajbe_decoder_init(&decoder, NULL, &reader.super);

        int64_t int_value;
        yajbe_decode_next_int(&decoder, &int_value);
        z_assert_eq(utest, ptest->value, int_value);
    }

    return 0;
}

int main (int argc, char **argv) {
    z_utest_t utest;

    z_utest_init(&utest);

    test_yajbe_int_simple(&utest);
    z_utest_dump_result(&utest);

    return 0;
}
