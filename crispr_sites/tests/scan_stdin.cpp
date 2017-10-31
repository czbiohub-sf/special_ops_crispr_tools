#include "catch.hpp"

#include <unistd.h>

#include <cstdlib>
#include <vector>
#include <iostream>
#include "../crispr_sites.hpp"

using namespace std;

// unit tests for scan_stdin()

void scan_stdin(bool output_counts);

char random_base(void) {
    switch (rand() % 4) {
    case 0:
	return 'A';
    case 1:
	return 'C';
    case 2:
	return 'G';
    case 3:
	return 'T';
    }
}

void random_kmer(char* output) {
    for (int i = 0; i < k; i++) {
	output[i] = random_base();
    }
}


TEST_CASE( "scan_stdin correctly finds crispr sites", "[scan_stdin]" ) {

    // hijack stdin and stdout
    int stdin_hijack[2];
    int original_stdout;

    original_stdout = dup(STDOUT_FILENO);

    FILE* output_file = fopen("scan_stdin_output.txt", "w");
    
    pipe(stdin_hijack);
    dup2(stdin_hijack[0], STDIN_FILENO);
    close(stdin_hijack[0]);
    
    //pipe(stdout_hijack);
    dup2(fileno(output_file), STDOUT_FILENO);    

    SECTION( "expected crispr sites detected" ) {
	constexpr auto INPUT_SIZE = 1024;

	char input[INPUT_SIZE];

	vector<string> expected_crispr_sites;
	
	for (int i = 0; i < INPUT_SIZE; i++) {
	    if ((i + k - 3) < INPUT_SIZE && rand() % 10 == 0) {
		// Add a crispr site
		if (rand() % 2) {
		    input[i++] = 'C';
		    input[i++] = 'C';
		    input[i++] = 'N';
		    random_kmer(&input[i]);
		    char crispr_site[k];
		    for (int j = 0; j < k; j++) {
			crispr_site[j] = input[i + j];
		    }
		    expected_crispr_sites.push_back(string(crispr_site));
		    
		    i += k - 1;
		} else {
		    random_kmer(&input[i]);
		    char crispr_site[k];
		    for (int j = 0; j < k; j++) {
			crispr_site[j] = input[i + j];
		    }
		    expected_crispr_sites.push_back(string(crispr_site));

		    i += k;
		    input[i++] = 'N';
		    input[i++] = 'G';
		    input[i] = 'G';
		}
	    } else {
		if (i > 0 && (input[i - 1] == 'C' || input[i - 1] == 'G')) {
		    // prevent random additional crispr sites from being genereated
		    input[i] = 'A'; 
		} else {
		    input[i] = random_base();
		}
	    }
	}

	FILE* sequence_output = fopen("scan_stdin_input.txt", "w");
	fwrite(input, sizeof(input), 1, sequence_output);
	fclose(sequence_output);

	write(stdin_hijack[1], input, INPUT_SIZE);
	close(stdin_hijack[1]);
	scan_stdin(false);
	cout.flush();
	fclose(output_file);

	dup2(original_stdout, STDOUT_FILENO);

	FILE *input_file = fopen("scan_stdin_output.txt", "r");

	int num_crispr_sites_found = 0;
	    
	char line[k + 1];
	while (fgets(line, k + 1, input_file) != NULL) {
	    line[k] = 0;
	    string crispr_site(line);

	    bool expected_crispr_site_found = false;
	    
	    for (auto it = expected_crispr_sites.begin(); it != expected_crispr_sites.end(); it++) {
		if (*it == crispr_site) {
		    expected_crispr_site_found = true;
		    break;
		}
	    }
	    if (!expected_crispr_site_found) {
		cerr << "Unexpected crispr site " << crispr_site << " found in output" << endl;
	    }
	    REQUIRE(expected_crispr_site_found == true);
	}

	REQUIRE(num_crispr_sites_found == expected_crispr_sites.size());
    }

}
