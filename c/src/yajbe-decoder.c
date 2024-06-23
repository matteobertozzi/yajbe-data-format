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
#include <errno.h>

#include "yajbe.h"
#include "util.h"

// NOTE: to avoid too many ifs, we pre-build a map with the tokens.
// so we can find the token just by looking up TOKEN_MAP[head]
static const int TOKEN_MAP[] = {
  0, 20, 1, 2,
  12, 13, 14, 15,
  8, 9, 9,
  -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
  16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 17,
  18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 18, 19,
  3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4,
  3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 5, 5, 5, 5, 5, 5, 5, 5,
  10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 11, 11, 11, 11,
  6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 7, 7, 7, 7
};

static int __read_item_count (yajbe_decoder_t *self, int64_t *count) {
    const int w = self->item_head & 0b1111;
    if (w <= 10) {
        *count = w;
        return 0;
    }

    int64_t value;
    const int r = z_bytes_reader_read_uint(self->reader, &value, w - 10);
    *count = 10 + value;
    return r;
}

static int __read_item_length (yajbe_decoder_t *self, int64_t *length) {
    int64_t value;
    const int r = z_bytes_reader_read_uint(self->reader, &value, (self->item_head & 0b111111) - 59);
    *length = 59 + value;
    return r;
}

void yajbe_decoder_init (yajbe_decoder_t *self, yajbe_field_decoder_t *field_reader, z_bytes_reader_t *reader) {
    self->reader = reader;
    self->field_reader = field_reader;
    self->item_head = -1;
    self->item_type = -1;
    self->item_length = 0;
}

int yajbe_decode_next (yajbe_decoder_t *self) {
    int r;

    r = z_bytes_reader_read_u8(self->reader, &(self->item_head));
    if (Z_UNLIKELY(r < 0)) return r;

    self->item_type = TOKEN_MAP[self->item_head];
    switch (self->item_type) {
        case YAJBE_ARRAY:
        case YAJBE_OBJECT:
            r = __read_item_count(self, &(self->item_length));
            if (Z_UNLIKELY(r < 0)) return r;
            break;
        case YAJBE_ARRAY_EOF:
        case YAJBE_OBJECT_EOF:
            self->item_length = 1L << 63;
            break;
        case YAJBE_SMALL_BYTES:
        case YAJBE_SMALL_STRING:
            self->item_length = self->item_head & 0b111111;
            break;
        case YAJBE_BYTES:
        case YAJBE_STRING:
            r = __read_item_length(self, &(self->item_length));
            if (Z_UNLIKELY(r < 0)) return r;
            break;
        case YAJBE_INT_SMALL:
            self->item_length = 0;
            break;
        case YAJBE_INT_POSITIVE:
        case YAJBE_INT_NEGATIVE:
            self->item_length = (self->item_head & 0b11111) - 23;
            break;
        case YAJBE_FLOAT_32:
            self->item_length = 4;
            break;
        case YAJBE_FLOAT_64:
            self->item_length = 8;
            break;
    }
    return self->item_type;
}

int yajbe_decode_null (yajbe_decoder_t *self) {
    return self->item_type != YAJBE_NULL;
}

int yajbe_decode_next_null (yajbe_decoder_t *self) {
    const int r = yajbe_decode_next(self);
    if (Z_LIKELY(r == YAJBE_NULL)) return 0;
    return r < 0 ? r : -EINVAL;
}

int yajbe_decode_true (yajbe_decoder_t *self) {
    return self->item_type != YAJBE_TRUE;
}

int yajbe_decode_next_true (yajbe_decoder_t *self) {
    const int r = yajbe_decode_next(self);
    if (Z_LIKELY(r == YAJBE_TRUE)) return 0;
    return r < 0 ? r : -EINVAL;
}

int yajbe_decode_false (yajbe_decoder_t *self) {
    return self->item_type != YAJBE_FALSE;
}

int yajbe_decode_next_false (yajbe_decoder_t *self) {
    const int r = yajbe_decode_next(self);
    if (Z_LIKELY(r == YAJBE_FALSE)) return 0;
    return r < 0 ? r : -EINVAL;
}

int yajbe_decode_bool (yajbe_decoder_t *self, int *value) {
    switch (self->item_type) {
        case YAJBE_TRUE:
            *value = 1;
            return 0;
        case YAJBE_FALSE:
            *value = 0;
            return 0;
    }
    return -EINVAL;
}

int yajbe_decode_next_bool (yajbe_decoder_t *self, int *value) {
    const int r = yajbe_decode_next(self);
    if (Z_UNLIKELY(r < 0)) return 0;
    return yajbe_decode_bool(self, value);
}

static int __decode_small_int (const yajbe_decoder_t *self, int64_t *value) {
    const int head = self->item_head;
    const int signed_value = (head & 0b01100000) == 0b01100000;
    const int w = head & 0b11111;
    *value = signed_value ? -w : (1 + w);
    return 0;
}

static int __decode_int_positive (yajbe_decoder_t *self, int64_t *value) {
  int r = z_bytes_reader_read_uint(self->reader, value, self->item_length);
  *value += 25L;
  return r;
}

static int __decode_int_negative (yajbe_decoder_t *self, int64_t *value) {
  int r = z_bytes_reader_read_uint(self->reader, value, self->item_length);
  *value = -(*value + 24L);
  return r;
}

int yajbe_decode_int (yajbe_decoder_t *self, int64_t *value) {
    switch (self->item_type) {
        case YAJBE_INT_SMALL:    return __decode_small_int(self, value);
        case YAJBE_INT_POSITIVE: return __decode_int_positive(self, value);
        case YAJBE_INT_NEGATIVE: return __decode_int_negative(self, value);
    }
    return -EINVAL;
}

int yajbe_decode_next_int (yajbe_decoder_t *self, int64_t *value) {
    const int r = yajbe_decode_next(self);
    if (Z_UNLIKELY(r < 0)) return 0;
    return yajbe_decode_int(self, value);
}

int yajbe_decode_float (yajbe_decoder_t *self, float *value) {
    if (Z_UNLIKELY(self->item_type != YAJBE_FLOAT_32)) {
        return -EINVAL;
    }
    int64_t iValue;
    int r = z_bytes_reader_read_uint(self->reader, &iValue, 4);
    memcpy(value, &iValue, 4);
    return r;
}

int yajbe_decode_next_float (yajbe_decoder_t *self, float *value) {
    const int r = yajbe_decode_next(self);
    if (Z_UNLIKELY(r < 0)) return 0;
    return yajbe_decode_float(self, value);
}

int yajbe_decode_double (yajbe_decoder_t *self, double *value) {
    if (Z_UNLIKELY(self->item_type != YAJBE_FLOAT_64)) {
        return -EINVAL;
    }
    int64_t iValue;
    int r = z_bytes_reader_read_uint(self->reader, &iValue, 8);
    memcpy(value, &iValue, 8);
    return r;
}

int yajbe_decode_next_double (yajbe_decoder_t *self, double *value) {
    const int r = yajbe_decode_next(self);
    if (Z_UNLIKELY(r < 0)) return 0;
    return yajbe_decode_double(self, value);
}

int yajbe_decode_bytes (yajbe_decoder_t *self, void *buf, size_t buflen) {
    if (Z_UNLIKELY((self->item_type != YAJBE_SMALL_BYTES && self->item_type != YAJBE_BYTES)
        || buflen < self->item_length
    )) {
        return -EINVAL;
    }
    return z_bytes_reader_read(self->reader, buf, self->item_length);
}

int yajbe_decode_next_bytes (yajbe_decoder_t *self, void *buf, size_t buflen) {
    const int r = yajbe_decode_next(self);
    if (Z_UNLIKELY(r < 0)) return 0;
    return yajbe_decode_bytes(self, buf, buflen);
}

int yajbe_decode_string (yajbe_decoder_t *self, char *buf, size_t buflen) {
    if (Z_UNLIKELY((self->item_type != YAJBE_SMALL_STRING && self->item_type != YAJBE_STRING)
        || buflen < self->item_length
    )) {
        return -EINVAL;
    }
    return z_bytes_reader_read(self->reader, buf, self->item_length);
}

int yajbe_decode_next_string (yajbe_decoder_t *self, char *buf, size_t buflen) {
    const int r = yajbe_decode_next(self);
    if (Z_UNLIKELY(r < 0)) return 0;
    return yajbe_decode_string(self, buf, buflen);
}

int yajbe_decode_object_field (yajbe_decoder_t *self, yajbe_field_t **field) {
    return yajbe_field_decode(self->field_reader, self->reader, field);
}
