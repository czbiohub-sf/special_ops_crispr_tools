CPPFLAGS=--std=c++11 -O3

crispr_sites : crispr_sites.o
	g++ $(CPPFLAGS) -o crispr_sites crispr_sites.o

crispr_sites.o : crispr_sites.cpp crispr_sites.hpp
	g++ $(CPPFLAGS) -c crispr_sites.cpp
