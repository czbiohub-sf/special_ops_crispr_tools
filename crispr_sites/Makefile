PROGRAM_NAME=crispr_sites
PROGRAM_VERSION := $(shell git describe --dirty --always --tags)

CPPFLAGS=--std=c++11 -O3

$(PROGRAM_NAME) : crispr_sites.o
	g++ $(CPPFLAGS) -o crispr_sites crispr_sites.o

crispr_sites.o : crispr_sites.cpp crispr_sites.hpp
	g++ $(CPPFLAGS) -DPROGRAM_VERSION=\"$(PROGRAM_VERSION)\" -DPROGRAM_NAME=\"$(PROGRAM_NAME)\" -c crispr_sites.cpp

.PHONY: clean

clean:
	rm -f $(PROGRAM_NAME) crispr_sites.o

