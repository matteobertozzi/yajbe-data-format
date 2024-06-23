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

#ifndef __YAJBE_H__
#define __YAJBE_H__

#include <stdint.h>
#include <stdlib.h>

// ================================================================================
//  Bytes Writer
// ================================================================================
typedef struct z_bytes_writer z_bytes_writer_t;

typedef struct {
  int (*write_u8)    (z_bytes_writer_t *self, int value);
  int (*write_uint)  (z_bytes_writer_t *self, int64_t value, int width);
  int (*write_bytes) (z_bytes_writer_t *self, const void *buf, size_t length);
} z_bytes_writer_vtable_t;

struct z_bytes_writer {
    const z_bytes_writer_vtable_t *vtable;
};

#define z_bytes_writer_init(self, impl_vtable)          (self)->vtable = impl_vtable;
#define z_bytes_writer_write(self, buf, len)            (self)->vtable->write_bytes(self, buf, len);
#define z_bytes_writer_write_u8(self, v)                (self)->vtable->write_u8(self, v);
#define z_bytes_writer_write_uint(self, v, vlen)        (self)->vtable->write_uint(self, v, vlen);

// ================================================================================
//  Bytes Writer
// ================================================================================
typedef struct {
    const uint8_t *buffer;
    size_t length;
} z_bytes_slice_t;

#define z_bytes_slice_set(self, xbuf, xlen)     \
    do {                                        \
        (self)->buffer = xbuf;                  \
        (self)->length = xlen;                  \
    } while (0)

void z_bytes_slice_dump_hex (const z_bytes_slice_t *self);
void z_bytes_slice_dump_string (const z_bytes_slice_t *self);

typedef struct z_bytes_reader z_bytes_reader_t;

typedef struct {
  int (*read_u8)    (z_bytes_reader_t *self, int *value);
  int (*read_uint)  (z_bytes_reader_t *self, int64_t *value, int width);
  int (*read_bytes) (z_bytes_reader_t *self, void *buf, size_t length);
  int (*read_slice) (z_bytes_reader_t *self, z_bytes_slice_t *slice, size_t length);
} z_bytes_reader_vtable_t;

struct z_bytes_reader {
    const z_bytes_reader_vtable_t *vtable;
};

#define z_bytes_reader_init(self, impl_vtable)          (self)->vtable = impl_vtable;
#define z_bytes_reader_read(self, buf, len)             (self)->vtable->read_bytes(self, buf, len);
#define z_bytes_reader_read_u8(self, v)                 (self)->vtable->read_u8(self, v);
#define z_bytes_reader_read_uint(self, vptr, vlen)      (self)->vtable->read_uint(self, vptr, vlen);
#define z_bytes_reader_read_slice(self, slice, len)     (self)->vtable->read_slice(self, slice, len);

// ================================================================================
//  InMemory Bytes Writer
// ================================================================================
typedef struct {
  z_bytes_writer_t super;
  uint8_t *buffer;
  size_t   bufsize;
  size_t   offset;
} z_mem_bytes_writer_t;

void z_mem_bytes_writer_init_fixed (z_mem_bytes_writer_t *self, uint8_t *buffer, size_t bufsize);

#define z_mem_bytes_writer_length(self)           ((self)->offset)

// ================================================================================
//  InMemory Bytes Reader
// ================================================================================
typedef struct {
  z_bytes_reader_t super;
  const uint8_t *buffer;
  size_t   bufsize;
  size_t   offset;
} z_mem_bytes_reader_t;

void z_mem_bytes_reader_init (z_mem_bytes_reader_t *self, const uint8_t *buffer, size_t bufsize);

// ================================================================================
//  Field Name Decoder
// ================================================================================
typedef struct {
    const char *name;
    size_t length;
    uint32_t hash;
    int index;
} yajbe_field_encoder_entry_t;

typedef struct  {
   yajbe_field_encoder_entry_t *entries;
   size_t entries_size;
   size_t entries_count;

   const char *last_key;
   size_t last_keylen;
} yajbe_field_encoder_t;

void yajbe_field_encoder_init_fixed (yajbe_field_encoder_t *self, yajbe_field_encoder_entry_t *entries, size_t entries_size);

int yajbe_field_encoder_hadd (yajbe_field_encoder_t *self, uint32_t khash, const char *key, size_t keylen);
int yajbe_field_encoder_hget (const yajbe_field_encoder_t *self, uint32_t khash, const char *key, size_t keylen);

int yajbe_field_encoder_add (yajbe_field_encoder_t *self, const char *key, size_t keylen);
int yajbe_field_encoder_get (const yajbe_field_encoder_t *self, const char *key, size_t keylen);

int yajbe_field_hencode (yajbe_field_encoder_t *self, z_bytes_writer_t *writer, uint32_t khash, const char *key, size_t keylen);
int yajbe_field_encode (yajbe_field_encoder_t *self, z_bytes_writer_t *writer, const char *key, size_t keylen);

// ================================================================================
//  YAJBE Encoder
// ================================================================================
typedef struct {
    z_bytes_writer_t *writer;
    yajbe_field_encoder_t *field_writer;
} yajbe_encoder_t;

void yajbe_encoder_init (yajbe_encoder_t *self, yajbe_field_encoder_t *field_writer, z_bytes_writer_t *writer);

int yajbe_encode_null (yajbe_encoder_t *self);
int yajbe_encode_true (yajbe_encoder_t *self);
int yajbe_encode_false (yajbe_encoder_t *self);
int yajbe_encode_bool (yajbe_encoder_t *self, int value);
int yajbe_encode_int (yajbe_encoder_t *self, int64_t value);
int yajbe_encode_float (yajbe_encoder_t *self, float value);
int yajbe_encode_double (yajbe_encoder_t *self, double value);
int yajbe_encode_bytes (yajbe_encoder_t *self, const void *buf, size_t length);
int yajbe_encode_string (yajbe_encoder_t *self, const char *utf8, size_t length);

int yajbe_encode_array_fixed_length(yajbe_encoder_t *self, size_t length);
int yajbe_encode_array_start (yajbe_encoder_t *self);
int yajbe_encode_array_end (yajbe_encoder_t *self);

int yajbe_encode_object_fixed_length(yajbe_encoder_t *self, size_t length);
int yajbe_encode_object_start (yajbe_encoder_t *self);
int yajbe_encode_object_end (yajbe_encoder_t *self);

int yajbe_encode_object_hfield (yajbe_encoder_t *self, uint32_t khash, const char *key, size_t keylen);
int yajbe_encode_object_field (yajbe_encoder_t *self, const char *key, size_t keylen);

// ================================================================================
//  Field Name Decoder
// ================================================================================
typedef struct {
    const char *name;
    size_t length;
} yajbe_field_t;

typedef struct  {
    yajbe_field_t *entries;
    size_t entries_size;
    size_t entries_count;

    char *buffer;
    size_t buf_size;
    size_t buf_off;

    yajbe_field_t last_field;
} yajbe_field_decoder_t;

void yajbe_field_decoder_init_fixed (yajbe_field_decoder_t *self, yajbe_field_t *entries, size_t entries_size, char *buffer, size_t bufsize);
int yajbe_field_decode (yajbe_field_decoder_t *self, z_bytes_reader_t *reader, yajbe_field_t **field);
void yajbe_field_dump (const yajbe_field_t *self);

// ================================================================================
//  YAJBE Encoder
// ================================================================================
typedef enum {
    YAJBE_NULL         = 0,
    YAJBE_FALSE        = 1,
    YAJBE_TRUE         = 2,
    YAJBE_INT_SMALL    = 3,
    YAJBE_INT_POSITIVE = 4,
    YAJBE_INT_NEGATIVE = 5,
    YAJBE_SMALL_STRING = 6,
    YAJBE_STRING       = 7,
    YAJBE_ENUM_CONFIG  = 8,
    YAJBE_ENUM_STRING  = 9,
    YAJBE_SMALL_BYTES  = 10,
    YAJBE_BYTES        = 11,
    YAJBE_FLOAT_VLE    = 12,
    YAJBE_FLOAT_32     = 13,
    YAJBE_FLOAT_64     = 14,
    YAJBE_BIG_DECIMAL  = 15,
    YAJBE_ARRAY        = 16,
    YAJBE_ARRAY_EOF    = 17,
    YAJBE_OBJECT       = 18,
    YAJBE_OBJECT_EOF   = 19,
    YAJBE_EOF          = 20,
} yajbe_item_type_t;

typedef struct {
    z_bytes_reader_t *reader;
    yajbe_field_decoder_t *field_reader;

    int item_head;
    int item_type;
    int64_t item_length;
} yajbe_decoder_t;

void yajbe_decoder_init (yajbe_decoder_t *self, yajbe_field_decoder_t *field_reader, z_bytes_reader_t *reader);

int yajbe_decode_next (yajbe_decoder_t *self);
#define yajbe_decode_next_item_type(self)             ((self)->item_type)
#define yajbe_decode_next_item_length(self)           ((self)->item_length)

int yajbe_decode_null (yajbe_decoder_t *self);
int yajbe_decode_bool (yajbe_decoder_t *self, int *value);
int yajbe_decode_int (yajbe_decoder_t *self, int64_t *value);
int yajbe_decode_float (yajbe_decoder_t *self, float *value);
int yajbe_decode_double (yajbe_decoder_t *self, double *value);
int yajbe_decode_bytes (yajbe_decoder_t *self, void *buf, size_t buflen);
int yajbe_decode_string (yajbe_decoder_t *self, char *buf, size_t buflen);
int yajbe_decode_object_field (yajbe_decoder_t *self, yajbe_field_t **field);

int yajbe_decode_next_null (yajbe_decoder_t *self);
int yajbe_decode_next_bool (yajbe_decoder_t *self, int *value);
int yajbe_decode_next_int (yajbe_decoder_t *self, int64_t *value);
int yajbe_decode_next_float (yajbe_decoder_t *self, float *value);
int yajbe_decode_next_double (yajbe_decoder_t *self, double *value);
int yajbe_decode_next_bytes (yajbe_decoder_t *self, void *buf, size_t buflen);
int yajbe_decode_next_string (yajbe_decoder_t *self, char *buf, size_t buflen);

#endif /* !__YAJBE_H__ */
