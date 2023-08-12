#!/bin/bash
scicmd='python ../scicmd.py -c'

# Create a fasta file with some DNA
$scicmd 'echo AAAGCCCGTGGGGGACCTGTTC > o:dna.fa'

# Compute the complement sequence
$scicmd 'cat i:dna.fa | tr ACGT TGCA > o:dna.compl.fa'

# Reverse the DNA string
$scicmd 'cat i:dna.compl.fa | rev > o:dna.compl.rev.fa'
