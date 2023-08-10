#!/usr/bin/env python
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

    print(f"Now executing command: {command} ...")
    out = sub.run(
        command, shell=True, stdout=sub.PIPE, stderr=sub.PIPE, text=True, check=True
    )

    if out.stdout:
        print(f"OUTPUT: {out.stdout}")
    if out.stderr:
        print(f"ERRORS: {out.stderr}")


def get_input_matches(text):
    pattern = re.compile("(\{i\:([^\{\}]+)\})")
    matches = pattern.findall(text)
    return matches


def get_output_matches(text):
    pattern = re.compile("(\{o\:([^\{\}]+)\}|o\:([^\s]+))")
    matches = pattern.findall(text)
    outputs = []
    for m in matches:
        output = tuple(m[0])
        if m[1]:
            output = m[0], m[1]
        elif m[2]:
            output = m[0], m[2]
        outputs.append(output)
    return outputs


if __name__ == "__main__":
    main()
