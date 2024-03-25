#!/bin/bash

# Create a fasta file with some DNA
sci run echo AAAGCCCGTGGGGGACCTGTTC '>' dna.fa

# Compute the complement sequence
sci run cat dna.fa '|' tr ACGT TGCA '>' dna.compl.fa

# Reverse the DNA string
sci run cat dna.compl.fa '|' rev '>' dna.compl.rev.fa
