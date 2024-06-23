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

#include "yajbe.h"
#include "util.h"

void z_bytes_slice_dump_hex (const z_bytes_slice_t *self) {
    printf("%zu:0x", self->length);
    for (size_t i = 0; i < self->length; ++i) {
        printf("%02x", self->buffer[i]);
    }
}

void z_bytes_slice_dump_string (const z_bytes_slice_t *self) {
    printf("%zu:", self->length);
    for (size_t i = 0; i < self->length; ++i) {
        printf("%c", self->buffer[i]);
    }
}
