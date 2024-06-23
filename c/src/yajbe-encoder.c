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

#include "yajbe.h"
#include "util.h"

#define __int_bytes_width(value)    (((value) != 0) ? ((64 - __builtin_clzll(value)) + 7) >> 3 : 1)

void yajbe_encoder_init (yajbe_encoder_t *self, yajbe_field_encoder_t *field_writer, z_bytes_writer_t *writer) {
    self->field_writer = field_writer;
    self->writer = writer;
}

int yajbe_encode_null (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 0);
}

int yajbe_encode_true (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 0b11);
}

int yajbe_encode_false (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 0b10);
}

int yajbe_encode_bool (yajbe_encoder_t *self, int value) {
    return z_bytes_writer_write_u8(self->writer, value ? 0b11 : 0b10);
}

static int __yajbe_encode_length (yajbe_encoder_t *self, int head, int inline_max, size_t length) {
    if (length <= inline_max) {
        return z_bytes_writer_write_u8(self->writer, head | length);
    }

    const size_t delta_length = length - inline_max;
    const int bytes = __int_bytes_width(delta_length);
    int r = z_bytes_writer_write_u8(self->writer, head | (inline_max + bytes));
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write_uint(self->writer, delta_length, bytes);
}

int yajbe_encode_array_fixed_length(yajbe_encoder_t *self, size_t length) {
    return __yajbe_encode_length(self, 0b00100000, 10, length);
}

int yajbe_encode_array_start (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 0b00101111);
}

int yajbe_encode_array_end (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 1);
}

int yajbe_encode_object_fixed_length(yajbe_encoder_t *self, size_t length) {
    return __yajbe_encode_length(self, 0b00110000, 10, length);
}

int yajbe_encode_object_start (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 0b00111111);
}

int yajbe_encode_object_end (yajbe_encoder_t *self) {
    return z_bytes_writer_write_u8(self->writer, 1);
}

int yajbe_encode_object_hfield (yajbe_encoder_t *self, uint32_t khash, const char *key, size_t keylen) {
    return yajbe_field_hencode(self->field_writer, self->writer, khash, key, keylen);
}

int yajbe_encode_object_field (yajbe_encoder_t *self, const char *key, size_t keylen) {
    return yajbe_field_encode(self->field_writer, self->writer, key, keylen);
}

static int __yajbe_encode_positive_int (yajbe_encoder_t *self, int64_t value) {
    if (value <= 24) {
      return z_bytes_writer_write_u8(self->writer, 0b01000000 | (value - 1));
    }

    value -= 25;
    const int bytes = __int_bytes_width(value);
    int r = z_bytes_writer_write_u8(self->writer, 0b01000000 | (23 + bytes));
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write_uint(self->writer, value, bytes);
}

static int __yajbe_encode_negative_int (yajbe_encoder_t *self, int64_t value) {
    value = -value;
    if (value <= 23) {
        return z_bytes_writer_write_u8(self->writer, 0b01100000 | value);
    }
    value -= 24;
    const int bytes = __int_bytes_width(value);
    int r = z_bytes_writer_write_u8(self->writer, 0b01100000 | (23 + bytes));
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write_uint(self->writer, value, bytes);
}

int yajbe_encode_int (yajbe_encoder_t *self, int64_t value) {
    if (value > 0) {
        return __yajbe_encode_positive_int(self, value);
    }
    return __yajbe_encode_negative_int(self, value);
}

int yajbe_encode_float (yajbe_encoder_t *self, float value) {
    int r = z_bytes_writer_write_u8(self->writer, 0b00000101);
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write(self->writer, &value, 4);
}

int yajbe_encode_double (yajbe_encoder_t *self, double value) {
    int r = z_bytes_writer_write_u8(self->writer, 0b00000110);
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write(self->writer, &value, 8);
}

int yajbe_encode_string (yajbe_encoder_t *self, const char *utf8, size_t length) {
    int r = __yajbe_encode_length(self, 0b11000000, 59, length);
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write(self->writer, utf8, length);
}

int yajbe_encode_bytes (yajbe_encoder_t *self, const void *buf, size_t length) {
    int r = __yajbe_encode_length(self, 0b10000000, 59, length);
    if (Z_UNLIKELY(r < 0)) return r;
    return z_bytes_writer_write(self->writer, buf, length);
}
