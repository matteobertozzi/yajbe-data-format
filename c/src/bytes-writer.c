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
#include <errno.h>

#include "yajbe.h"
#include "util.h"

// ================================================================================
//  InMemory Bytes Writer
// ================================================================================
static int __mem_bytes_writer_write_byte (z_bytes_writer_t *super, int v) {
    z_mem_bytes_writer_t *self = (z_mem_bytes_writer_t *)super;
    if (Z_UNLIKELY(self->bufsize < (self->offset + 1))) {
        return -ENOSPC;
    }
    self->buffer[self->offset++] = v & 0xff;
    return 0;
}

static int __mem_bytes_writer_write_uint (z_bytes_writer_t *super, int64_t value, int width) {
    z_mem_bytes_writer_t *self = (z_mem_bytes_writer_t *)super;
    const size_t off = self->offset;

    if (Z_UNLIKELY(self->bufsize < (self->offset + width))) {
        return -ENOSPC;
    }

    uint8_t *buf = self->buffer;
    switch (width) {
      case 8: buf[off + 7] = (value >> 56) & 0xff;
      case 7: buf[off + 6] = (value >> 48) & 0xff;
      case 6: buf[off + 5] = (value >> 40) & 0xff;
      case 5: buf[off + 4] = (value >> 32) & 0xff;
      case 4: buf[off + 3] = (value >> 24) & 0xff;
      case 3: buf[off + 2] = (value >> 16) & 0xff;
      case 2: buf[off + 1] = (value >> 8) & 0xff;
      case 1: buf[off] = value & 0xff;
    }
    self->offset += width;
    return 0;
}

static int __mem_bytes_writer_write_bytes (z_bytes_writer_t *super, const void *buf, size_t length) {
    z_mem_bytes_writer_t *self = (z_mem_bytes_writer_t *)super;
    if (Z_UNLIKELY(self->bufsize < (self->offset + length))) {
        return -ENOSPC;
    }

    memcpy(self->buffer + self->offset, buf, length);
    self->offset += length;
    return 0;
}

static const z_bytes_writer_vtable_t __mem_bytes_writer_vtable = {
    .write_u8       = __mem_bytes_writer_write_byte,
    .write_uint     = __mem_bytes_writer_write_uint,
    .write_bytes    = __mem_bytes_writer_write_bytes,
};

void z_mem_bytes_writer_init_fixed (z_mem_bytes_writer_t *self, uint8_t *buffer, size_t bufsize) {
    z_bytes_writer_init(&(self->super), &__mem_bytes_writer_vtable);
    self->buffer = buffer;
    self->bufsize = bufsize;
    self->offset = 0;
}
