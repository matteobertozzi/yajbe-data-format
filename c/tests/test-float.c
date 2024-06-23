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
#include <sys/_types/_null.h>

#include "yajbe.h"
#include "utest.h"

struct test_float_value {
    float value;
    uint8_t buffer[5];
};

static const struct test_float_value test_values[] = {
    { .value = 0.0f, .buffer = { 0x05, 0x00, 0x00, 0x00, 0x00 } },
    { .value = 1.0f, .buffer = { 0x05, 0x00, 0x00, 0x80, 0x3f } },
    { .value = 1.1f, .buffer = { 0x05, 0xcd, 0xcc, 0x8c, 0x3f }},
    { .value = -32.26664f, .buffer = { 0x05, 0x0a, 0x11, 0x01, 0xc2 }},
};

static int test_yajbe_float_simple (z_utest_t *utest) {
    z_mem_bytes_writer_t writer;
    yajbe_encoder_t encoder;
    uint8_t buffer[32];

    z_mem_bytes_writer_init_fixed(&writer, buffer, sizeof(buffer));
    yajbe_encoder_init(&encoder, NULL, &writer.super);

    for (int i = 0, n = sizeof(test_values) / sizeof(struct test_float_value); i < n; ++i) {
        const struct test_float_value *ptest = test_values + i;
        writer.offset = 0;
        yajbe_encode_float(&encoder, ptest->value);
        z_assert_eq(utest, 5, writer.offset);
        z_assert_eq_data(utest, ptest->buffer, 5, buffer, 5);
    }

    return 0;
}

static int test_yajbe_float_encode_decode (z_utest_t *utest) {
    z_mem_bytes_writer_t writer;
    z_mem_bytes_reader_t reader;
    yajbe_encoder_t encoder;
    yajbe_decoder_t decoder;
    uint8_t buffer[32];

    z_mem_bytes_writer_init_fixed(&writer, buffer, sizeof(buffer));
    yajbe_encoder_init(&encoder, NULL, &writer.super);

    for (int i = 0, n = sizeof(test_values) / sizeof(struct test_float_value); i < n; ++i) {
        const struct test_float_value *ptest = test_values + i;

        // encode
        writer.offset = 0;
        yajbe_encode_float(&encoder, ptest->value);
        z_assert_eq(utest, 5, writer.offset);
        z_assert_eq_data(utest, ptest->buffer, 5, buffer, 5);

        // decode
        z_mem_bytes_reader_init(&reader, buffer, writer.offset);
        yajbe_decoder_init(&decoder, NULL, &reader.super);

        float float_value;
        yajbe_decode_next_float(&decoder, &float_value);
        z_assert_eq_float(utest, ptest->value, float_value, 0.000001f);
    }

    return 0;
}

int main (int argc, char **argv) {
    z_utest_t utest;

    z_utest_init(&utest);

    test_yajbe_float_simple(&utest);
    z_utest_dump_result(&utest);

    test_yajbe_float_encode_decode(&utest);
    z_utest_dump_result(&utest);

    return 0;
}
