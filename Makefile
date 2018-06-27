.PHONY: build

NAME = mbmi-go
VERSION = $(shell cat VERSION | sed -e 's,\-.*,,')
RELEASE = $(shell cat VERSION | sed -e 's,.*\-,,')

BUILD_DIR = $(notdir $(shell pwd))
BUILD_DATE = $(shell date +%Y%m%d%H%M%S)
BUILD_ARCH = amd64

LDFLAGS = -ldflags "-X main.programName=${NAME} -X main.programVersion=${VERSION} -X main.buildDate=${BUILD_DATE}"

SOURCE_FILES = $(shell ls -AB | grep -i 'version$$\|makefile$$\|\.go$$')

# Debian build root
DEB_DIR = $(shell pwd)/build/debian
DEB_ROOT = $(DEB_DIR)/$(NAME)-$(VERSION)/debian
DEB_CONF = $(DEB_ROOT)/$(NAME).conf
DEB_INITD = $(DEB_ROOT)/$(NAME).init.sh
DEB_SOURCE = $(DEB_DIR)/$(NAME)_$(VERSION).orig.tar.gz
DEB_PKG = $(DEB_DIR)/$(NAME)_$(VERSION)-$(RELEASE)_$(BUILD_ARCH).deb

build:
	go build -o ./$(NAME) -v $(LDFLAGS)

cover:
	go test -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html

test:
	@go test -v .

bench:
	@go test -bench=. -benchmem

dependency:
	@go get -fix -t $(BUILD_PKGS)

deb: .clean_deb $(DEB_PKG)

$(DEB_ROOT): contrib/debian
	mkdir -p $(DEB_ROOT)
	cp -ad $</* $@/
	find $@ -type f -exec sed -i -e"s/@VERSION@/$(VERSION)/g" {} \;

$(DEB_SOURCE): $(SOURCE_FILES)
	mkdir -p $(@D)
	tar --transform "s,^,$(NAME)-$(VERSION)/src/$(NAME)/," -f $@ -cz $^

$(DEB_CONF): contrib/conf/$(NAME).conf
	mkdir -p $(@D)
	cp -ad $< $@

$(DEB_INITD): contrib/scripts/$(NAME).init.sh
	mkdir -p $(@D)
	cp -ad $< $@

$(DEB_PKG): $(DEB_ROOT) $(DEB_SOURCE) $(DEB_CONF) $(DEB_INITD)
	cd $(DEB_DIR)/$(NAME)-$(VERSION) && \
	debuild --set-envvar BUILD_APP_VERSION=$(VERSION) -us -uc -b

.clean_deb:
	@rm -rf $(shell find . -type d -path "*build/debian*" -print -quit)

