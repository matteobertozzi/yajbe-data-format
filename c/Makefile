CC = gcc
CFLAGS = -Wall -std=c11 -g -I./include
#CFLAGS = -Wall -Wextra -std=c11
LDFLAGS = -L./build -lyajbe

SRC_DIR = src
TEST_DIR = tests
EXAMPLES_DIR = examples
INCLUDE_DIR = include
BUILD_DIR = build
OBJ_DIR = $(BUILD_DIR)/obj

SOURCES = $(wildcard $(SRC_DIR)/*.c)
OBJECTS = $(patsubst $(SRC_DIR)/%.c, $(OBJ_DIR)/%.o, $(SOURCES))

TESTS = $(wildcard $(TEST_DIR)/*.c)
TEST_EXECUTABLES = $(patsubst $(TEST_DIR)/%.c, $(TEST_DIR)/%, $(TESTS))

EXAMPLESS = $(wildcard $(EXAMPLES_DIR)/*.c)
EXAMPLES_EXECUTABLES = $(patsubst $(EXAMPLES_DIR)/%.c, $(EXAMPLES_DIR)/%, $(EXAMPLESS))

# Define the name of the library
LIBRARY = $(BUILD_DIR)/libyajbe.a

# Default target
all: $(LIBRARY) $(TEST_EXECUTABLES) $(EXAMPLES_EXECUTABLES)

# Create the static library
$(LIBRARY): $(OBJECTS)
	$(AR) rcs $@ $^

# Create the object directory if it doesn't exist
$(OBJ_DIR):
	mkdir -p $(OBJ_DIR)

# Compile Library
$(OBJ_DIR)/%.o: $(SRC_DIR)/%.c | $(OBJ_DIR)
	$(CC) $(CFLAGS) -c $< -o $@

# Compile Tests
$(TEST_DIR)/%: $(TEST_DIR)/%.c $(LIBRARY)
	$(CC) $(CFLAGS) $< -o $@ $(LDFLAGS)

# Compile Examples
$(EXAMPLES_DIR)/%: $(EXAMPLES_DIR)/%.c $(LIBRARY)
	$(CC) $(CFLAGS) $< -o $@ $(LDFLAGS)

# Run all tests
test: $(TEST_EXECUTABLES)
	@for test in $(TEST_EXECUTABLES); do \
		echo "Running $$test..."; \
		./$$test; \
	done

examples: $(EXAMPLES_EXECUTABLES)
	@for examples in $(EXAMPLES_EXECUTABLES); do \
		echo "Running $$examples..."; \
		./$$examples; \
	done

# Clean up build files
clean:
	rm -rf $(OBJ_DIR) $(LIBRARY) $(TEST_EXECUTABLES) $(EXAMPLES_EXECUTABLES)
	find . -iname '*.dSYM' | xargs rm -r

# Phony targets
.PHONY: all clean test
