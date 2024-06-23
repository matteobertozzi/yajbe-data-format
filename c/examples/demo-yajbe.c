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

#include <stdint.h>
#include <stdio.h>
#include "yajbe.h"

#define yajbe_encode_array(self, code_block)        \
    do {                                            \
        int r;                                      \
        r = yajbe_encode_array_start(self);         \
        if (r < 0) return r;                        \
        do code_block while (0);                    \
        r = yajbe_encode_array_end(self);           \
        if (r < 0) return r;                        \
    } while (0)

#define yajbe_encode_object(self, code_block)       \
    do {                                            \
        int r;                                      \
        r = yajbe_encode_object_start(self);        \
        if (r < 0) return r;                        \
        do code_block while (0);                    \
        r = yajbe_encode_object_end(self);          \
        if (r < 0) return r;                        \
    } while (0)

static void dump_hex (const uint8_t *buf, size_t bufsize) {
    printf("%ld: ", bufsize);
    for (size_t i = 0; i < bufsize; ++i) {
        printf("%02x", buf[i]);
    }
    printf("\n");
}

static size_t demo_encode (uint8_t *buffer, size_t bufsize) {
    yajbe_field_encoder_entry_t field_entries[16];
    yajbe_field_encoder_t field_encoder;
    z_mem_bytes_writer_t writer;
    yajbe_encoder_t encoder;

    z_mem_bytes_writer_init_fixed(&writer, buffer, bufsize);
    yajbe_field_encoder_init_fixed(&field_encoder, field_entries, 16);
    yajbe_encoder_init(&encoder, &field_encoder, &(writer.super));

    yajbe_encode_array_fixed_length(&encoder, 1);
    yajbe_encode_object(&encoder, {
        yajbe_encode_object_field(&encoder, "field_null", 10);
        yajbe_encode_null(&encoder);
        yajbe_encode_object_field(&encoder, "bool_true", 9);
        yajbe_encode_true(&encoder);
        yajbe_encode_object_field(&encoder, "bool_false", 10);
        yajbe_encode_false(&encoder);
        yajbe_encode_object_field(&encoder, "field_int_0", 11);
        yajbe_encode_int(&encoder, 3);
        yajbe_encode_object_field(&encoder, "field_int_1", 11);
        yajbe_encode_int(&encoder, 1234);
        yajbe_encode_object_field(&encoder, "field_int_2", 11);
        yajbe_encode_int(&encoder, -543210);
        yajbe_encode_object_field(&encoder, "field_sm_str", 12);
        yajbe_encode_string(&encoder, "foo", 3);
    });

    return z_mem_bytes_writer_length(&writer);
}

static const char *YAJBE_ITEM_TYPE[] = {
    "YAJBE_NULL",
    "YAJBE_FALSE", "YAJBE_TRUE",
    "YAJBE_INT_SMALL", "YAJBE_INT_POSITIVE", "YAJBE_INT_NEGATIVE",
    "YAJBE_SMALL_STRING", "YAJBE_STRING",
    "YAJBE_ENUM_CONFIG", "YAJBE_ENUM_STRING",
    "YAJBE_SMALL_BYTES", "YAJBE_BYTES",
    "YAJBE_FLOAT_VLE", "YAJBE_FLOAT_32", "YAJBE_FLOAT_64", "YAJBE_BIG_DECIMAL",
    "YAJBE_ARRAY", "YAJBE_ARRAY_EOF",
    "YAJBE_OBJECT", "YAJBE_OBJECT_EOF",
    "YAJBE_EOF",
};

static void demo_decode (uint8_t *buffer, size_t bufsize) {
    yajbe_field_decoder_t field_decoder;
    yajbe_field_t field_entries[16];
    z_mem_bytes_reader_t reader;
    yajbe_decoder_t decoder;
    yajbe_field_t *field;
    char buf_fields[256];

    z_mem_bytes_reader_init(&reader, buffer, bufsize);
    yajbe_field_decoder_init_fixed(&field_decoder, field_entries, 16, buf_fields, sizeof(buf_fields));
    yajbe_decoder_init(&decoder, &field_decoder, &(reader.super));

    yajbe_decode_next(&decoder); // expected array length:1
    printf("%s:%lld []\n", YAJBE_ITEM_TYPE[decoder.item_type], decoder.item_length);

    yajbe_decode_next(&decoder); // expected object eof
    printf("%s:%lld {}\n", YAJBE_ITEM_TYPE[decoder.item_type], decoder.item_length);

    yajbe_decode_object_field(&decoder, &field);
    yajbe_decode_next_null(&decoder);
    yajbe_field_dump(field); printf(" = NULL\n");

    for (int i = 0; i < 2; ++i) {
        int bool_value;
        yajbe_decode_object_field(&decoder, &field);
        yajbe_field_dump(field);
        yajbe_decode_next_bool(&decoder, &bool_value);
        yajbe_field_dump(field); printf(" = BOOL(%d)\n", bool_value);
    }

    for (int i = 0; i < 3; ++i) {
        int64_t int_value;
        yajbe_decode_object_field(&decoder, &field);
        yajbe_field_dump(field);
        yajbe_decode_next_int(&decoder, &int_value);
        yajbe_field_dump(field); printf(" = INT(%lld)\n", int_value);
    }

    if (1) {
        char str_buf[64];
        yajbe_decode_object_field(&decoder, &field);
        yajbe_field_dump(field);
        yajbe_decode_next_string(&decoder, str_buf, sizeof(str_buf));
        str_buf[decoder.item_length] = '\0';
        yajbe_field_dump(field); printf(" = STR(%lld:%s)\n", decoder.item_length, str_buf);
    }
}

int main (int argc, char **argv) {
    uint8_t buffer[1024];
    size_t enclen;

    enclen = demo_encode(buffer, sizeof(buffer));
    dump_hex(buffer, enclen);
    demo_decode(buffer, enclen);

    return 0;
}
