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
#include <string.h>
#include <stdio.h>
#include <errno.h>

#include "yajbe.h"
#include "util.h"

static void __field_decoder_set_last_key (yajbe_field_decoder_t *self, const yajbe_field_t *field) {
    self->last_field.name = field->name;
    self->last_field.length = field->length;
}

static yajbe_field_t *__field_next_entry (yajbe_field_decoder_t *self, size_t length) {
    if (Z_UNLIKELY(self->buf_size < (self->buf_off + length))) {
        return NULL;
    }

    yajbe_field_t *field = &(self->entries[self->entries_count++]);
    field->name = self->buffer + self->buf_off;
    field->length = length;
    self->buf_off += length;
    return field;
}

void yajbe_field_set (yajbe_field_t *self, const char *name, size_t length) {
    self->name = name;
    self->length = length;
}

void yajbe_field_decoder_init_fixed (yajbe_field_decoder_t *self, yajbe_field_t *entries, size_t entries_size, char *buffer, size_t bufsize) {
    self->entries = entries;
    self->entries_size = entries_size;
    self->entries_count = 0;

    self->buffer = buffer;
    self->buf_size = bufsize;
    self->buf_off = 0;

    self->last_field.name = NULL;
    self->last_field.length = 0;
}

static int __read_length (z_bytes_reader_t *reader, const int head) {
  const int length = (head & 0b00011111);
  if (length < 30) return length;
  if (length == 30) {
      int v;
      z_bytes_reader_read_u8(reader, &v);
      return 29 + v;
  }

  uint8_t buf[2];
  z_bytes_reader_read(reader, buf, 2);
  return 284 + 256 * (buf[0] & 0xff) + (buf[1] & 0xff);
}


static int __read_full_field_name (yajbe_field_decoder_t *self, z_bytes_reader_t *reader, const int head, yajbe_field_t **field) {
    if (Z_UNLIKELY(self->entries_count == self->entries_size)) {
        return -ENOSPC;
    }

    const int length = __read_length(reader, head);
    *field = __field_next_entry(self, length);
    if (Z_UNLIKELY(*field == NULL)) {
        return -ENOSPC;
    }

    char *name = (char *) ((*field)->name);
    z_bytes_reader_read(reader, name, length);

    __field_decoder_set_last_key(self, *field);
    return 0;
}


static int __read_indexed_field_name (yajbe_field_decoder_t *self, z_bytes_reader_t *reader, const int head, yajbe_field_t **field) {
    const int index = __read_length(reader, head);
    *field = &(self->entries[index]);
    __field_decoder_set_last_key(self, *field);
    return 0;
}

static int __read_prefix (yajbe_field_decoder_t *self, z_bytes_reader_t *reader, const int head, yajbe_field_t **field) {
    if (Z_UNLIKELY(self->entries_count == self->entries_size)) {
        return -ENOSPC;
    }

    int prefix, length;
    length = __read_length(reader, head);
    z_bytes_reader_read_u8(reader, &prefix);

    *field = __field_next_entry(self, prefix + length);
    if (Z_UNLIKELY(*field == NULL)) {
        return -ENOSPC;
    }

    char *name = (char *) ((*field)->name);
    memcpy(name, self->last_field.name, prefix); name += prefix;
    z_bytes_reader_read(reader, name, length);

    __field_decoder_set_last_key(self, *field);
    return 0;
}

static int __read_prefix_suffix (yajbe_field_decoder_t *self, z_bytes_reader_t *reader, const int head, yajbe_field_t **field) {
    if (Z_UNLIKELY(self->entries_count == self->entries_size)) {
        return -ENOSPC;
    }

    uint8_t delta[2];
    int length;

    length = __read_length(reader, head);
    z_bytes_reader_read(reader, delta, 2);

    *field = __field_next_entry(self, delta[0] + delta[1] + length);
    if (Z_UNLIKELY(*field == NULL)) {
        return -ENOSPC;
    }

    char *name = (char *) ((*field)->name);
    memcpy(name, self->last_field.name, delta[0]); name += delta[0];
    z_bytes_reader_read(reader, name, length); name += length;
    memcpy(name, self->last_field.name + self->last_field.length - delta[1], delta[1]);

    __field_decoder_set_last_key(self, *field);
    return 0;
}

int yajbe_field_decode (yajbe_field_decoder_t *self, z_bytes_reader_t *reader, yajbe_field_t **field) {
    int head;
    z_bytes_reader_read_u8(reader, &head);
    switch ((head >> 5) & 0b111) {
      case 0b100: return __read_full_field_name(self, reader, head, field);
      case 0b101: return __read_indexed_field_name(self, reader, head, field);
      case 0b110: return __read_prefix(self, reader, head, field);
      case 0b111: return __read_prefix_suffix(self, reader, head, field);
    }
    fprintf(stderr, "unexpected head: %02x", head);
    return -1;
}

void yajbe_field_dump (const yajbe_field_t *self) {
    printf("%ld:", self->length);
    for (size_t i = 0; i < self->length; ++i) {
        printf("%c", self->name[i]);
    }
}