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

#include <string.h>
#include <stdint.h>
#include <stdio.h>

#include "yajbe.h"
#include "utest.h"

static const char *TEST_FIELDS[] = {
    "foo",
    "bar",
    "test_foo",
    "test_bar",
    "foo",
    "prefix_foo_suffix",
    "prefix_bar_suffix",
    "bar",
    "test_foo",
    NULL,
};

static int encode_fields (uint8_t *buffer, size_t bufsize) {
    yajbe_field_encoder_entry_t entries[16];
    yajbe_field_encoder_t field_encoder;
    z_mem_bytes_writer_t writer;

    z_mem_bytes_writer_init_fixed(&writer, buffer, bufsize);

    yajbe_field_encoder_init_fixed(&field_encoder, entries, 16);
    for (int i = 0; TEST_FIELDS[i] != NULL; ++i) {
        const char *key = TEST_FIELDS[i];
        size_t keylen = strlen(key);
        //printf(" - ENCODE %zd:%s\n", keylen, key);
        yajbe_field_encode(&field_encoder, &writer.super, key, keylen);
    }

    return writer.offset;
}

static int decode_fields (z_utest_t *utest, const uint8_t *buffer, size_t bufsize) {
    yajbe_field_decoder_t field_decoder;
    z_mem_bytes_reader_t reader;
    yajbe_field_t entries[16];
    yajbe_field_t *field;
    char buf_fields[256];

    z_mem_bytes_reader_init(&reader, buffer, bufsize);

    yajbe_field_decoder_init_fixed(&field_decoder, entries, 16, buf_fields, sizeof(buf_fields));
    for (int i = 0; TEST_FIELDS[i] != NULL; ++i) {
        const char *key = TEST_FIELDS[i];
        yajbe_field_decode(&field_decoder, &reader.super, &field);
        //printf(" - DECODE ");
        z_assert_eq_data(utest, key, strlen(key), field->name, field->length);
        //yajbe_field_dump(field);
        //printf(" (%s)\n", key);
    }
    return 0;
}

int main (int argc, char **argv) {
    z_bytes_slice_t slice;
    uint8_t buffer[64];
    z_utest_t utest;

    z_utest_init(&utest);

    int encode_len = encode_fields(buffer, sizeof(buffer));

    z_bytes_slice_set(&slice, buffer, encode_len);
    z_bytes_slice_dump_hex(&slice);

    decode_fields(&utest, buffer, encode_len);
    z_utest_dump_result(&utest);
    return 0;
}
