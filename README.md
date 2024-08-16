SciCommander
============

[![Build and test Go code](https://github.com/samuell/scicommander/actions/workflows/go-ci.yml/badge.svg)](https://github.com/samuell/scicommander/actions/workflows/go-ci.yml)

This is a small tool that executes single shell commands in a scientifically
more reproducible and robust way, by doing the following things:

## Features

- Auditing: Creating an audit log of most output files
- Caching: Skipping executions where output files already exist

## Roadmap

There are also some further features that are planned to be introduced further
down the road, such as:

- Atomic writes - Writes files to a temporary location (such as a sub-folder)
  until command is finished, so that they are never placed at their final
  destination before completely finished.

## Current status

- **May 2, 2024**: I'm currently rewriting this tool in Go, to allow easier
  distribution on multiple platforms, and avoiding to lock the tool into one
  specific python environment. The core functionality is already ported, and
  also heavily improved. Only the HTML reporting remains to port.

## News

- **Aug 16, 2024:** Version 0.4 released, which is a complete rewrite of the
  tool in Go, with some cool new features such as the ability to detect output
  files automatically.
- **Nov 9, 2023:** Version 0.3.3 released, with a new command, `scishell`, that
  allows you to run commands more like in a normal shell (only adding  `i:`
  before input files and `o:` before output files), instead of running it
  through a separate command, and still have the full audit trace generated.
  - Read more in [this blog post](https://bionics.it/posts/scicommander-0.3)!

## Requirements

- A unix like operating system such as Linux or Mac OS (On Windows you can use
  [WSL](https://learn.microsoft.com/en-us/windows/wsl/about) or [MSYS2](https://www.msys2.org/))
- A Bash shell
- For graph plotting for the HTML report, you need
  [GraphViz](https://graphviz.org/) and its `dot` command.

## Installation

### Using Go

This method assumes that you have [installed the Go toolchain](https://go.dev/doc/install).

```bash
go install github.com/samuell/scicommander@latest
```

This will install the `sci` command into your `PATH` variable, so that it
should be executable from your shell.

(Other installation options to be added shortly)

## Usage

To view the options of the `sci` command, execute:

```bash
sci -h
```

To get the benefits from SciCommander, do the following:

1. Prepend all your shell commands with the `sci run` command.
2. Wrap the command itself in quotes, either `""` or `''`. This is not strictly
   required always, but will be required for example if using redirection using
   `>` or piping with `|` (Alternatively one can just add quotes around those).
3. Then run your script as usual.

Now you will notice that if you run your script again, it will skip all
commands that have already finished and produced output files.

You will also have files with the extension `.au` for every output that you
decorated with the syntax above.

To convert such an audit report into a nice HTML-report, you can run the
following:

```bash
sci to-html <audit-file>
```

## Example

To demonstrate how you can use SciCommander, imagine that you want to write the
following little toy bioinformatics pipeline, that writes some DNA and converts
its reverse complement, as a shell script, `my_pipeline.sh`:

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
sci run echo AAAGCCCGTGGGGGACCTGTTC '>' o:dna.fa
# Compute the complement sequence
sci run cat i:dna.fa '|' tr ACGT TGCA '>' dna.compl.fa
# Reverse the DNA string
sci run cat dna.compl.fa '|' rev '>' dna.compl.rev.fa
```

Notice that we had to wrap all pipe characters (`|`) and redirection characters
(`>`) in quotes. This is so that they are not grabbed by bash immediately but
instead passed with the command to SciCommander, and executed as part of its
execution.

An alternative is to encapsulate the full commands in `''`:

```bash
#!/bin/bash

# Create a fasta file with some DNA
sci run 'echo AAAGCCCGTGGGGGACCTGTTC > o:dna.fa'
# Compute the complement sequence
sci run 'cat i:dna.fa | tr ACGT TGCA > dna.compl.fa'
# Reverse the DNA string
sci run 'cat dna.compl.fa | rev > dna.compl.rev.fa'
```

Now you can run the script as usual, e.g. with:

```bash
bash my_pipeline.sh
```

Now, the files in your folder will look like this, if you list them with `ls -tr`:

```bash
my_pipeline.sh
dna.fa.au
dna.fa
dna.compl.fa.au
dna.compl.fa
dna.compl.rev.fa.au
dna.compl.rev.fa
```

Now, you see that the last `.au` file is `dna.compl.rev.fa.au`.

To convert this file to HTML and view it in a browser, you can do:

```bash
sci to-html dna.compl.rev.fa.au
```

## Experimental: Bash integration

There is experimental support for running SciCommander commands in bash,
without needing to run them via the `sci run` command.

To do this, start the SciCommander shell with the following command:

```bash
sci shell
```

And then, you can run the example commands above as follows:

```bash
# Create a fasta file with some DNA
echo AAAGCCCGTGGGGGACCTGTTC > o:dna.fa
# Compute the complement sequence
cat dna.fa | tr ACGT TGCA > o:dna.compl.fa
# Reverse the DNA string
cat dna.compl.fa | rev > o:dna.compl.rev.fa
```

In other words, no extra syntax is needed.

## Notes

[1] Although Nextflow and Snakemake already take care of some of the benefits,
such as atomic writes, SciCommander adds additional features such as detailed
per-output audit logs. It can thus be a great complement to these tools.
