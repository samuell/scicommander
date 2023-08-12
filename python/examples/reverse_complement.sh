#!/bin/bash

# Create a fasta file with some DNA
scicmd -c echo AAAGCCCGTGGGGGACCTGTTC '>' o:dna.fa

# Compute the complement sequence
scicmd -c cat i:dna.fa '|' tr ACGT TGCA '>' o:dna.compl.fa

# Reverse the DNA string
scicmd -c cat i:dna.compl.fa '|' rev '>' o:dna.compl.rev.fa
