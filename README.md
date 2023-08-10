SciCommander
============

This is a small tool that executes single shell commands in a scientifically
more reproducible and robust way, by doing the following things:

- Atomic writes: Writes files to a temporary location until command is finished
- Auditing: Creating an audit log of most output files
- Caching: Skipping executions where output files already exist

It allows executing shell commands with very minor modifications:

1. Prepend your command with the `scicmd` command.
2. Wrap definitions of input fields in `{i:INPATH}` and output files in
   `{o:OUTPATH}` for output paths.

The benefits are multiple:

1. Detailed audit logs for every command (Compatible with SciPipe audit logs).
2. Unfinished output files are never written to final paths
3. Avoid re-running already completed commands
4. Can be combined with any existing scripting solution, such as:
  - Shell scripts
  - Python scripts
  - Nextflow pipelines [1]
  - Snakemake pipelines [1]

## Notes

[1] Although Nextflow and Snakemake already take care of some of the benefits,
such as atomic writes, SciExec adds additional features such as detailed
per-output audit logs.
