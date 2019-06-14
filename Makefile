prefix ?= /usr/local

setup: build-all 

build-all: build-cripsr-sites build-offtarget

build-cripsr-sites:
	cd crispr_sites && $(MAKE)

build-offtarget:
	cd offtarget && go build -o offtarget

.PHONY: test

test:
	cd crispr_sites && make tests

install: build-cripsr-sites build-offtarget
	install crispr_sites/crispr_sites $(prefix)/bin
	install offtarget/offtarget $(prefix)/bin
