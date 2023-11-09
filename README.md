SciCommander
============

[![Run Python Tests](https://github.com/samuell/scicommander/actions/workflows/python-run-tests.yml/badge.svg)](https://github.com/samuell/scicommander/actions/workflows/python-run-tests.yml)

This is a small tool that executes single shell commands in a scientifically
more reproducible and robust way, by doing the following things:

- Auditing: Creating an audit log of most output files
- Caching: Skipping executions where output files already exist
- (Coming soon): Atomic writes - Writes files to a temporary location until
  command is finished

## Requirements

- A unix like operating system such as Linux or Mac OS (On Windows you can use
  [WSL](https://learn.microsoft.com/en-us/windows/wsl/about) or [MSYS2](https://www.msys2.org/))
- Python 3.6 or higher
- A bash shell
- For graph plotting for the HTML report, you need
  [GraphViz](https://graphviz.org/) and its `dot` command.

## Installation

```bash
pip install scicommander
```

This will install the `scicmd` command into your `PATH` variable, so that it
should be executable from your shell.

## Usage

To view the options of the `scicmd` command, execute:

```bash
scicmd -h
```

To get the benefits from SciCommander, do the following:

1. Prepend all your shell commands with the `scicmd -c` command.
2. Wrap the command itself in quotes, either `""` or `''`. This is not strictly
   required always, but will be required for example if using redirection using
   `>` or piping with `|` (Alternatively one can just add quotes around those).
3. Wrap definitions of input fields in `{i:INPATH}` and output files in
   `{o:OUTPATH}` for output paths.
4. You can also just prepend input paths with `i:` and output paths with `o:`,
   but this is a slightly less robust method, that might fail to wrap the
   correct number of characters in some situations.
5. Then run your script as usual.

Now you will notice that if you run your script again, it will skip all
commands that have already finished and produced output files.

You will also have files with the extension `.au.json` for every output that
you decorated with the syntax above.

To convert such an audit report into a nice HTML-report, run the following:

```bash
scicmd --to-html <audit-file>

```

## Example

To demonstrate how you can use SciCommander, imagine that you want to write the
following little bioinformatics pipeline, that writes some DNA and converts its
reverse complement, as a shell script, `my_pipeline.sh`:

```bash
#!/bin/bash

# Create a fasta file with some DNA
echo AAAGCCCGTGGGGGACCTGTTC > o:dna.fa
# Compute the complement sequence
cat i:dna.fa | tr ACGT TGCA > o:dna.compl.fa
# Reverse the DNA string
cat i:dna.compl.fa | rev > o:dna.compl.rev.fa
```
Now, to make the commands run through SciCommander, change the syntax in the
script like this:

```bash
#!/bin/bash

# Create a fasta file with some DNA
scicmd -c echo AAAGCCCGTGGGGGACCTGTTC '>' o:dna.fa
# Compute the complement sequence
scicmd -c cat i:dna.fa '|' tr ACGT TGCA '>' o:dna.compl.fa
# Reverse the DNA string
scicmd -c cat i:dna.compl.fa '|' rev '>' o:dna.compl.rev.fa
```

Notice how all input paths are prepended with `i:` and output paths with `o:`,
and also that we had to wrap all pipe characters (`|`) and redirection
characters (`>`) in quotes. This is so that they are not grabbed by bash
immediately, but instead passed with the command to SciCommander, and executed
as part of its execution.

Now you can run the script as usual, e.g. with:

```bash
bash my_pipeline.sh
```

Now, the files in your folder will look like this, if you list them with `ls -tr`:

```bash
my_pipeline.sh
dna.fa.au.json
dna.fa
dna.compl.fa.au.json
dna.compl.fa
dna.compl.rev.fa.au.json
dna.compl.rev.fa
```

Now, you see that the last `.au.json` file is `dna.compl.rev.fa.au.json`.

To convert this file to HTML and view it in a browser, you can do:

```bash
scicmd --to-html dna.compl.rev.fa.au.json
```

Then you will see [an HTML page like this](https://htmlpreview.github.io/?https://github.com/samuell/scicommander/blob/main/python/examples/dna.compl.rev.fa.au.html)

## Experimental: Bash integration

There is very early and experimental support for running SciCommander commands
in bash, without needing to run them via the `scicmd -c` command.

To do this, start the SciCommander shell with the following command:

```bash
scishell
```

And then, you can run the example commands above as follows:

```bash
# Create a fasta file with some DNA
echo AAAGCCCGTGGGGGACCTGTTC > o:dna.fa
# Compute the complement sequence
cat i:dna.fa | tr ACGT TGCA > o:dna.compl.fa
# Reverse the DNA string
cat i:dna.compl.fa | rev > o:dna.compl.rev.fa
```

In other words, only the `i:` and `o:` markers are now needed, and no extra
syntax.

## Notes

[1] Although Nextflow and Snakemake already take care of some of the benefits,
such as atomic writes, SciCommander adds additional features such as detailed
per-output audit logs.
