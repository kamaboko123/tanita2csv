# Makefile

CMD_SRC_DIR := cmd
LIB_SRC_DIR := pkg
TARGET_DIR := bin
# Find directories containing main.go files and the directory name to be used as target(executable file) names
# e.g., cmd/hoge/main.go -> hoge, cmd/fuga/main.go -> fuga
EXECUTABLE_FILENAMES := $(shell find $(CMD_SRC_DIR) -name main.go | xargs dirname | xargs basename)
TARGETS := $(addprefix $(TARGET_DIR)/, $(EXECUTABLE_FILENAMES))
LIB_SRC_FILES := $(shell find $(LIB_SRC_DIR) -name '*.go')

# リリース関連の設定
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
RELEASE_DIR := releases
PLATFORMS := linux/amd64 windows/amd64

# リリース用のビルド設定
define build-release
	GOOS=$(word 1,$(subst /, ,$(1))) GOARCH=$(word 2,$(subst /, ,$(1))) go build \
		-ldflags "-s -w -X main.version=$(VERSION)" \
		-o $(RELEASE_DIR)/$(2)-$(VERSION)-$(subst /,-,$(1))$(if $(findstring windows,$(1)),.exe,) \
		$(CMD_SRC_DIR)/$(2)/main.go
endef

.PHONY: all
all: $(TARGETS)

$(TARGET_DIR)/%: $(CMD_SRC_DIR)/%/main.go $(LIB_SRC_FILES)
	mkdir -p $(TARGET_DIR)
	go build -o $@ $<

# リリース用のターゲット
.PHONY: release
release: clean-release
	mkdir -p $(RELEASE_DIR)
	$(foreach platform,$(PLATFORMS),$(foreach target,$(EXECUTABLE_FILENAMES),$(call build-release,$(platform),$(target));))

.PHONY: release-all
release-all: release

# 新しいプラットフォーム用のリリースを追加する例
# PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: clean
clean:
	rm -rf $(TARGET_DIR)

.PHONY: clean-release
clean-release:
	rm -rf $(RELEASE_DIR)
