PREFIX ?= /usr/local

setup: build-all

build-all: build-cripsr-sites build-offtarget

build-cripsr-sites:
	cd crispr_sites && $(MAKE)

offtarget/offtarget: offtarget/main.go offtarget/matcher.go
	cd offtarget && go build -o offtarget

.PHONY: test

test:
	cd crispr_sites && make tests

install: build-cripsr-sites offtarget/offtarget
	install crispr_sites/crispr_sites $(PREFIX)/bin
	install offtarget/offtarget $(PREFIX)/bin
