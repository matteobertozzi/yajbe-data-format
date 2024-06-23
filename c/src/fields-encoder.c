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
#include <errno.h>

#include "yajbe.h"
#include "util.h"

static void __field_name_entry_set (yajbe_field_encoder_entry_t *self, int index, uint32_t hash, const char *key, const size_t keylen) {
    self->name = key;
    self->length = keylen;
    self->hash = hash;
    self->index = index;
}

static int __field_name_entry_equals (const yajbe_field_encoder_entry_t *self, uint32_t hash, const char *key, const size_t keylen) {
    return self->hash == hash
        && self->length == keylen
        && !memcmp(self->name, key, keylen);
}

static uint32_t __hash_fnv_1a (const char *key, size_t length) {
    uint32_t hash = 0x811c9dc5;
    for (size_t i = 0; i < length; ++i) {
        hash ^= key[i];
        hash *= 0x811C9DC5;
    }
    return hash;
}

void yajbe_field_encoder_init_fixed (yajbe_field_encoder_t *self, yajbe_field_encoder_entry_t *entries, size_t entries_size) {
    self->entries = entries;
    self->entries_size = entries_size;
    self->entries_count = 0;
    memset(entries, 0, entries_size * sizeof(yajbe_field_encoder_entry_t));

    self->last_key = NULL;
    self->last_keylen = 0;
}

int yajbe_field_encoder_hadd (yajbe_field_encoder_t *self, uint32_t khash, const char *key, size_t keylen) {
    if (Z_UNLIKELY(self->entries_count == self->entries_size)) {
        return -ENOSPC;
    }

    size_t mask = (self->entries_size - 1);

    size_t hindex = khash & mask;
    while (self->entries[hindex].name != NULL) {
        const yajbe_field_encoder_entry_t *entry = self->entries + hindex;
        if (__field_name_entry_equals(entry, khash, key, keylen)) {
            return entry->index;
        }
        hindex = (hindex + 1) & mask;
    }

    const size_t index = self->entries_count++;
    __field_name_entry_set(self->entries + hindex, index, khash, key, keylen);
    return index;
}

int yajbe_field_encoder_hget (const yajbe_field_encoder_t *self, uint32_t khash, const char *key, size_t keylen) {
    size_t mask = (self->entries_size - 1);

    size_t hindex = khash & mask;
    for (size_t i = 0; i < self->entries_size; ++i) {
        const yajbe_field_encoder_entry_t *entry = self->entries + hindex;
        if (entry->name == NULL) {
            return -1;
        } else if (__field_name_entry_equals(entry, khash, key, keylen)) {
            return entry->index;
        }
        hindex = (hindex + 1) & mask;
    }
    return -1;
}

int yajbe_field_encoder_add (yajbe_field_encoder_t *self, const char *key, size_t keylen) {
    size_t khash = __hash_fnv_1a(key, keylen);
    return yajbe_field_encoder_hadd(self, khash, key, keylen);
}

int yajbe_field_encoder_get (const yajbe_field_encoder_t *self, const char *key, size_t keylen) {
    size_t khash = __hash_fnv_1a(key, keylen);
    return yajbe_field_encoder_hget(self, khash, key, keylen);
}

static int __prefix (const yajbe_field_encoder_t *self, const char *key, size_t keylen) {
    const char *last_key = self->last_key;
    const size_t len = Z_MIN(self->last_keylen, keylen);
    for (size_t i = 0; i < len; ++i) {
        if (last_key[i] != key[i]) {
            return i;
        }
    }
    return len;
}

static int __suffix (const yajbe_field_encoder_t *self, const char *key, int prefix, size_t keylen) {
    const char *last_key = self->last_key;
    const size_t last_keylen = self->last_keylen;
    const size_t len = Z_MIN(last_keylen, keylen);
    for (size_t i = 1; i < len; ++i) {
        if (last_key[last_keylen - i] != key[keylen - i]) {
            return i - 1;
        }
    }
    return len;
}

static int __write_length (z_bytes_writer_t *writer, int head, size_t length) {
    if (length < 30) {
        z_bytes_writer_write_u8(writer, head | length);
    } else if (length <= 284) {
      z_bytes_writer_write_u8(writer, head | 0x1e);
      z_bytes_writer_write_u8(writer, (length - 29) & 0xff);
    } else if (length <= 65819) {
      z_bytes_writer_write_u8(writer, head | 0x1f);
      z_bytes_writer_write_u8(writer, (length - 284) / 256);
      z_bytes_writer_write_u8(writer, (length - 284) & 255);
    } else {
      // throw UnsupportedError("unexpected too many field names: $length");
      return -1;
    }
    return 0;
}

static int __encode_indexed_field (z_bytes_writer_t *writer, int field_index) {
    // 101----- Field Offset (0-29 field, 30 1b-len, 31 2b-len)
    return __write_length(writer, 0xa0, field_index);
}

static int __encode_prefix (z_bytes_writer_t *writer, const char *key, size_t keylen, int prefix) {
    // 110----- Prefix (1byte prefix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    const size_t length = keylen - prefix;
    __write_length(writer, 0xc0, length);
    z_bytes_writer_write_u8(writer, prefix);
    z_bytes_writer_write(writer, key + prefix, length);
    return 0;
}

static int __encode_prefix_suffix (z_bytes_writer_t *writer, const char *key, size_t keylen, int prefix, int suffix) {
    // 111----- Prefix/Suffix (1byte prefix, 1byte suffix, 0-29 length - 1, 30 1b-len, 31 2b-len)
    const size_t length = keylen - prefix - suffix;
    __write_length(writer, 0xe0, length);
    z_bytes_writer_write_u8(writer, prefix);
    z_bytes_writer_write_u8(writer, suffix);
    z_bytes_writer_write(writer, key + prefix, keylen - suffix);
    return 0;
}

static int __encode_full_field_name (z_bytes_writer_t *writer, const char *key, size_t keylen) {
    // 100----- Full Field Name (0-29 length - 1, 30 1b-len, 31 2b-len)
    __write_length(writer, 0x80, keylen);
    z_bytes_writer_write(writer, key, keylen);
    return 0;
}

int yajbe_field_hencode (yajbe_field_encoder_t *self, z_bytes_writer_t *writer, uint32_t khash, const char *key, size_t keylen) {
    int index = yajbe_field_encoder_hget(self, khash, key, keylen);
    if (index >= 0) {
        self->last_key = key;
        self->last_keylen = keylen;
        return __encode_indexed_field(writer, index);
    }

    if (self->last_key != NULL && self->last_keylen > 4) {
        const int prefix = Z_MIN(0xff, __prefix(self, key, keylen));
        const int suffix = __suffix(self, key, prefix, keylen - prefix);

        if (suffix > 2) {
            __encode_prefix_suffix(writer, key, keylen, prefix, Z_MIN(0xff, suffix));
        } else if (prefix > 2) {
            __encode_prefix(writer, key, keylen, prefix);
        } else {
            __encode_full_field_name(writer, key, keylen);
        }
    } else {
        __encode_full_field_name(writer, key, keylen);
    }

    yajbe_field_encoder_hadd(self, khash, key, keylen);
    self->last_key = key;
    self->last_keylen = keylen;
    return 0;
}

int yajbe_field_encode (yajbe_field_encoder_t *self, z_bytes_writer_t *writer, const char *key, size_t keylen) {
    size_t khash = __hash_fnv_1a(key, keylen);
    return yajbe_field_hencode(self, writer, khash, key, keylen);
}
