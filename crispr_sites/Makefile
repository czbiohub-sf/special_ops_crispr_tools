PROGRAM_NAME = crispr_sites
PROGRAM_VERSION := $(shell git describe --dirty --always --tags)
CXX ?= g++

$(PROGRAM_NAME) : crispr_sites.o
	$(CXX) $(CPPFLAGS) --std=c++11 -o crispr_sites crispr_sites.o

crispr_sites.o : crispr_sites.cpp crispr_sites.hpp
	$(CXX) $(CPPFLAGS) --std=c++11 -DPROGRAM_VERSION=\"$(PROGRAM_VERSION)\" -DPROGRAM_NAME=\"$(PROGRAM_NAME)\" -c crispr_sites.cpp

tests:
	cd tests && make && ./tests_all

.PHONY: clean tests

clean:
	rm -f $(PROGRAM_NAME) crispr_sites.o
	cd tests && make clean
