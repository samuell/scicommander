#!/usr/bin/env python
import sys
import os
import re
import subprocess


def main():
    command = sys.argv[0]
    args = sys.argv[1:]

    for arg in args:
        if os.path.isfile(arg):
            print(f"FILE: {arg}")
        else:
            print(f"Arg: {arg}")

    print("-" * 80)

    args_str = " ".join(args)
    input_matches = get_input_matches(args_str)
    output_matches = get_output_matches(args_str)

    matches = input_matches + output_matches

    for placeholder, path in matches:
        print(f"Full placeholder: {placeholder}, Path: {path}")


def get_input_matches(text):
    pattern = re.compile("(\{i\:([^\{\}]+)\})")
    matches = pattern.findall(text)
    return matches

def get_output_matches(text):
    pattern = re.compile("(\{o\:([^\{\}]+)\})")
    matches = pattern.findall(text)
    return matches


if __name__ == "__main__":
    main()
