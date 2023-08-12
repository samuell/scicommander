SciCommander
============

This is a small tool that executes single shell commands in a scientifically
more reproducible and robust way, by doing the following things:

- Auditing: Creating an audit log of most output files
- Caching: Skipping executions where output files already exist
- (Coming soon): Atomic writes - Writes files to a temporary location until
  command is finished

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

## Notes

[1] Although Nextflow and Snakemake already take care of some of the benefits,
such as atomic writes, SciCommander adds additional features such as detailed
per-output audit logs.
