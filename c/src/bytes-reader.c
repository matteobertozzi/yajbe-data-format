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

// ================================================================================
//  InMemory Bytes Reader
// ================================================================================
static int __mem_bytes_reader_read_byte (z_bytes_reader_t *super, int *v) {
    z_mem_bytes_reader_t *self = (z_mem_bytes_reader_t *)super;
    if (Z_UNLIKELY(self->bufsize < (self->offset + 1))) {
        return -ENOSPC;
    }
    *v = self->buffer[self->offset++] & 0xff;
    return 0;
}

static int __mem_bytes_reader_read_uint (z_bytes_reader_t *super, int64_t *value, int width) {
    z_mem_bytes_reader_t *self = (z_mem_bytes_reader_t *)super;
    const size_t off = self->offset;

    if (Z_UNLIKELY(self->bufsize < (self->offset + width))) {
        return -ENOSPC;
    }

    const uint8_t *buf = self->buffer;
    int64_t result = 0;
    switch (width) {
      case 8: result |= (buf[off + 7] & 0xFFL) << 56;
      case 7: result |= (buf[off + 6] & 0xFFL) << 48;
      case 6: result |= (buf[off + 5] & 0xFFL) << 40;
      case 5: result |= (buf[off + 4] & 0xFFL) << 32;
      case 4: result |= (buf[off + 3] & 0xFFL) << 24;
      case 3: result |= (buf[off + 2] & 0xFFL) << 16;
      case 2: result |= (buf[off + 1] & 0xFFL) << 8;
      case 1: result |= buf[off] & 0xFFL;
    }
    *value = result;
    self->offset += width;
    return 0;
}

static int __mem_bytes_reader_read_bytes (z_bytes_reader_t *super, void *buf, size_t length) {
    z_mem_bytes_reader_t *self = (z_mem_bytes_reader_t *)super;
    if (Z_UNLIKELY(self->bufsize < (self->offset + length))) {
        return -ENOSPC;
    }

    memcpy(buf, self->buffer + self->offset, length);
    self->offset += length;
    return 0;
}

static int __mem_bytes_reader_read_slice (z_bytes_reader_t *super, z_bytes_slice_t *slice, size_t length) {
    z_mem_bytes_reader_t *self = (z_mem_bytes_reader_t *)super;
    if (Z_UNLIKELY(self->bufsize < (self->offset + length))) {
        return -ENOSPC;
    }

    slice->buffer = self->buffer + self->offset;
    slice->length = length;
    self->offset += length;
    return 0;
}

static const z_bytes_reader_vtable_t __mem_bytes_reader_vtable = {
    .read_u8       = __mem_bytes_reader_read_byte,
    .read_uint     = __mem_bytes_reader_read_uint,
    .read_bytes    = __mem_bytes_reader_read_bytes,
    .read_slice    = __mem_bytes_reader_read_slice,
};

void z_mem_bytes_reader_init (z_mem_bytes_reader_t *self, const uint8_t *buffer, size_t bufsize) {
    z_bytes_reader_init(&(self->super), &__mem_bytes_reader_vtable);
    self->buffer = buffer;
    self->bufsize = bufsize;
    self->offset = 0;
}
