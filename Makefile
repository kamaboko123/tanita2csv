# Makefile

CMD_SRC_DIR := cmd
LIB_SRC_DIR := pkg
TARGET_DIR := bin
# Find directories containing main.go files and the directory name to be used as target(executable file) names
# e.g., cmd/hoge/main.go -> hoge, cmd/fuga/main.go -> fuga
EXECUTABLE_FILENAMES := $(shell find $(CMD_SRC_DIR) -name main.go | xargs dirname | xargs basename)
TARGETS := $(addprefix $(TARGET_DIR)/, $(EXECUTABLE_FILENAMES))
LIB_SRC_FILES := $(shell find $(LIB_SRC_DIR) -name '*.go')


.PHONY: all
all: $(TARGETS)
	@echo $(TARGETS)

$(TARGET_DIR)/%: $(CMD_SRC_DIR)/%/main.go $(LIB_SRC_FILES)
	mkdir -p $(TARGET_DIR)
	go build -o $@ $<

.PHONY: clean
clean:
	rm -rf $(TARGET_DIR)
