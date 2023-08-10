#!/usr/bin/env python
import json
import sys
import os
import re
import subprocess as sub


def main():
    script = sys.argv[0]
    args = sys.argv[1:]
    args_str = " ".join(args)

    input_paths = []
    input_matches = get_input_matches(args_str)
    for ph, path in input_matches:
        input_paths.append(path)

    output_paths = []
    output_matches = get_output_matches(args_str)
    for ph, path in output_matches:
        output_paths.append(path)

    command = args_str

    for path in output_paths:
        if os.path.exists(path):
            print(f"Skipping: {command} (Output exists: {path})")
            return

    # Replace placeholders with only the path
    matches = input_matches + output_matches
    for ph, path in matches:
        command = command.replace(ph, path)

    print(f"Now executing command: {command} ...")
    out = sub.run(
        command, shell=True, stdout=sub.PIPE, stderr=sub.PIPE, text=True, check=True
    )

    audit_info = {
        "command": command,
        "inputs": input_paths,
        "outputs": output_paths,
        "upstream": {},
    }

    for path in input_paths:
        audit_path = f"{path}.audit.json"
        if os.path.exists(path):
            with open(audit_path) as audit_f:
                upstream_audit_info = json.load(audit_f)
            audit_info["upstream"][path] = upstream_audit_info

    for path in output_paths:
        audit_path = f"{path}.audit.json"
        with open(audit_path, "w") as audit_f:
            json.dump(audit_info, audit_f, indent=2)

    if out.stdout:
        print(f"OUTPUT: {out.stdout}")
    if out.stderr:
        print(f"ERRORS: {out.stderr}")


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
    pattern = re.compile("(\{o\:([^\{\}]+)\}|o\:([^\s]+))")
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
    main()
