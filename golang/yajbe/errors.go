// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package yajbe

import (
	"fmt"
	"reflect"
)

// Common error patterns used throughout YAJBE

// decodeTypeError creates a consistent error message for type conversion failures
func decodeTypeError(sourceType string, targetType string) error {
	return fmt.Errorf("cannot decode %s into %s", sourceType, targetType)
}

// decodeIntoError creates a specific error for decode failures with target type
func decodeIntoError(sourceType string, target any) error {
	return decodeTypeError(sourceType, getTypeName(target))
}

// invalidLengthError creates an error for invalid length values
func invalidLengthError(contextType string, length int) error {
	return fmt.Errorf("invalid %s length: %d", contextType, length)
}

// Common typed error constructors for better type safety and consistency

// Common decode errors
func boolDecodeError(target any) error {
	return decodeIntoError("bool", target)
}

func intDecodeError(target any) error {
	return decodeIntoError("int", target)
}

func float32DecodeError(target any) error {
	return decodeIntoError("float32", target)
}

func float64DecodeError(target any) error {
	return decodeIntoError("float64", target)
}

func stringDecodeError(target any) error {
	return decodeIntoError("string", target)
}

func bytesDecodeError(target any) error {
	return decodeIntoError("bytes", target)
}

func arrayDecodeError(target any) error {
	return decodeIntoError("array", target)
}

func eofArrayDecodeError(target any) error {
	return decodeIntoError("EOF array", target)
}

func objectDecodeError(target any) error {
	return decodeIntoError("object", target)
}

// Validation errors with context
func arrayLengthError(length int) error {
	return invalidLengthError("array", length)
}

func valueTooLargeError(valueType string) error {
	return fmt.Errorf("%s value too large for YAJBE", valueType)
}

// Enhanced type name function with better error context
func getTypeNameWithContext(v any, context string) string {
	if v == nil {
		return fmt.Sprintf("nil (in %s)", context)
	}
	
	// Use reflection to get more detailed type information
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		if t.Elem().Kind() == reflect.Interface {
			return fmt.Sprintf("*interface{} (in %s)", context)
		}
		return fmt.Sprintf("*%s (in %s)", t.Elem().String(), context)
	}
	
	return fmt.Sprintf("%s (in %s)", t.String(), context)
}

// Error wrapper for adding context to existing errors
func wrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// Operation-specific error constructors
func encodeError(operation string, valueType string, err error) error {
	return fmt.Errorf("failed to encode %s during %s operation: %w", valueType, operation, err)
}

func decodeError(operation string, targetType string, err error) error {
	return fmt.Errorf("failed to decode to %s during %s operation: %w", targetType, operation, err)
}