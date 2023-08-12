SciCommander
============

This is a small tool that executes single shell commands in a scientifically
more reproducible and robust way, by doing the following things:

- Auditing: Creating an audit log of most output files
- Caching: Skipping executions where output files already exist
- (Coming soon): Atomic writes - Writes files to a temporary location until
  command is finished

It allows executing shell commands with very minor modifications:

1. Prepend your command with the `scicmd -c` command.
2. Wrap definitions of input fields in `{i:INPATH}` and output files in
   `{o:OUTPATH}` for output paths.
3. You can also just prepend input paths with `i:` and output paths with `o:`,
   but this is a slightly less robust method, that might fail to wrap the
   correct number of characters in some situations.

The benefits are multiple:

1. Detailed audit log for every output.
1. Avoid re-running already completed commands
2. Can be combined with any existing scripting solution, such as:
  - Shell scripts
  - Python scripts
  - Most pipeline tools [1]

## Notes

[1] Although Nextflow and Snakemake already take care of some of the benefits,
such as atomic writes, SciCommander adds additional features such as detailed
per-output audit logs.
