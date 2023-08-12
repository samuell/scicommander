#!/usr/bin/env python
import argparse
import json
import sys
import os
import re
import subprocess as sub

# Define commandline arguments
argp = argparse.ArgumentParser()
argp.add_argument("--command", "-c", nargs="...", required=True)
argp.add_argument(
    "--merge-audit-files",
    "-m",
    action="store_true",
    help="Whether to merge in upstream audit files, to create fully self-contained audit trails",
)
args = argp.parse_args()


def main(args):
    command = " ".join(args.command)

    # Capture input paths
    input_paths = []
    input_matches = get_input_matches(command)
    for ph, path in input_matches:
        input_paths.append(path)

    # Capture output paths
    output_paths = []
    output_matches = get_output_matches(command)
    for ph, path in output_matches:
        output_paths.append(path)

    # Check if paths already exist
    for path in output_paths:
        if os.path.exists(path):
            print(f"Skipping: {command} (Output exists: {path})")
            return

    # Replace placeholders with only the path
    print(f"Command before replace: {command}")
    matches = input_matches + output_matches
    for ph, path in matches:
        print(f"Replacing {ph} with {path}")
        command = command.replace(ph, path)
    print(f"Command after replace: {command}")

    # Execute command
    stdout, stderr, retcode = execute_command(command)

    # Write AuditInfo file(s)
    write_audit_files(command, input_paths, output_paths, args.merge_audit_files)


def write_audit_files(command, input_paths, output_paths, merge_audit_files):
    audit_extension = ".au.json"

    audit_info = {
        "command": command,
        "inputs": input_paths,
        "outputs": output_paths,
        "upstream": {},
    }

    # Merge input audits into the final one
    if merge_audit_files:
        for path in input_paths:
            audit_path = f"{path}.au.json"
            if os.path.exists(path):
                with open(audit_path) as audit_f:
                    upstream_audit_info = json.load(audit_f)
                audit_info["upstream"][path] = upstream_audit_info

    for path in output_paths:
        audit_path = f"{path}.au.json"
        with open(audit_path, "w") as audit_f:
            json.dump(audit_info, audit_f, indent=2)


def execute_command(command):
    print(f"Executing: {command} ...")
    out = sub.run(
        command,
        shell=True,
        stdout=sub.PIPE,
        stderr=sub.PIPE,
        text=True,
        check=True,
    )
    if out.stdout:
        print(f"OUTPUT: {out.stdout}")
    if out.stderr:
        print(f"ERRORS: {out.stderr}")
    if out.returncode != 0:
        raise Exception(f"Command failed with returncode {out.returncode}: {command}")
    return out.stdout.strip(), out.stderr.strip(), out.returncode


def get_input_matches(text):
    pattern = re.compile("(\{i\:([^\{\}]+)\}|i\:([^\s]+))")
    matches = pattern.findall(text)
    inputs = []
    for m in matches:
        inp = tuple(m[0])
        if m[1]:
            inp = m[0], m[1]
        elif m[2]:
            inp = m[0], m[2]
        inputs.append(inp)
    return inputs


def get_output_matches(text):
    pattern = re.compile("(\{o\:([^\{\}]+)\}|o\:([^\ ]+))")
    matches = pattern.findall(text)
    outputs = []
    for m in matches:
        outp = tuple(m[0])
        if m[1]:
            outp = m[0], m[1]
        elif m[2]:
            outp = m[0], m[2]
        outputs.append(outp)
    return outputs


if __name__ == "__main__":
    main(args)
