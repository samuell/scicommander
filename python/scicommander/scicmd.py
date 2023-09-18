#!/usr/bin/env python
import argparse
import datetime as dt
import json
import sys
import os
import re
import subprocess as sub

# Define commandline arguments
argp = argparse.ArgumentParser()
argp.add_argument(
    "--merge-audit-files",
    "-m",
    action="store_true",
    help="Whether to merge in upstream audit files, to create fully self-contained audit trails",
)
argp.add_argument("--command", "-c", nargs="...")
argp.add_argument("--to-html", "-th", metavar="AUDIT_FILE")
args = argp.parse_args()


def main():
    if not args.command and not args.to_html:
        argp.print_usage()
        print("Either --command or --to-html must be specified!")
        return

    # Only write HTML report and finish
    if args.to_html:
        audit_path = args.to_html
        tasks = collect_audit_info(audit_path)
        write_html_report(tasks, audit_path)
        return

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
    matches = input_matches + output_matches
    for ph, path in matches:
        command = command.replace(ph, path)

    # Execute command
    start_time = dt.datetime.now()
    stdout, stderr, retcode = execute_command(command)
    end_time = dt.datetime.now()

    # Write AuditInfo file(s)
    write_audit_files(
        command,
        input_paths,
        output_paths,
        start_time,
        end_time,
        args.merge_audit_files,
    )


def generate_dot_graph(tasks):
    nodes, edges = generate_graph(tasks)

    dot = "DIGRAPH G {\n"
    dot += "  node [shape=box, style=filled, fillcolor=lightgrey, fontname=monospace, penwidth=0];"
    for node in nodes:
        dot += f' "{node}"\n'
    for edge in edges:
        dot += f'  "{edge[0]}" -> "{edge[1]}"\n'
    dot += "}"

    return dot


def generate_graph(tasks):
    nodes = []
    edges = []
    for task in tasks:
        command = " ".join(task["executors"][0]["command"])
        nodes.append(command)
        for out_info in task["outputs"]:
            edges.append((command, out_info["url"]))
            nodes.append(out_info["url"])
        for in_info in task["inputs"]:
            edges.append((in_info["url"], command))
            nodes.append(in_info["url"])
    nodes = set(nodes)
    edges = set(edges)
    return nodes, edges


def collect_audit_info(audit_path):
    with open(audit_path) as aifile:
        ai = json.load(aifile)
    tasks = []

    def add_input_audit_files(ai):
        for in_info in ai["inputs"]:
            input_audit_path = f"{in_info['url']}.au.json"
            if not os.path.isfile(input_audit_path):
                return

            with open(input_audit_path) as iaupath:
                upstream = json.load(iaupath)
            tasks.append(upstream)
            add_input_audit_files(upstream)

    tasks.append(ai)
    add_input_audit_files(ai)
    tasks.sort(key=lambda x: x["tags"]["start_time"])

    # Remove duplicate tasks
    seen_commands = set()
    unique_tasks = []
    for task in tasks:
        command = " ".join(task["executors"][0]["command"])
        if command not in seen_commands:
            seen_commands.add(command)
            unique_tasks.append(task)

    return unique_tasks


def write_html_report(tasks, audit_path):
    dot = generate_dot_graph(tasks)
    dot_path = audit_path.replace(".au.json", ".au.dot")
    svg_path = dot_path.replace(".dot", ".svg")

    with open(dot_path, "w") as dotfile:
        dotfile.write(dot)

    sub.run(f"dot -Tsvg {dot_path} > {svg_path}", shell=True)

    with open(svg_path) as svg_file:
        svg = svg_file.read().strip()

    html = "<html>\n"
    html += "<body style='font-family:monospace, courier new'>\n"
    html += "<h1>SciCommander Audit Report<h1>\n"
    html += "<hr>\n"
    html += "<table borders='none' cellpadding='8px'>\n"
    html += "<tr><th>Start time</th><th>Command</th><th>Duration</th></tr>\n"
    for task in tasks:
        command = " ".join(task["executors"][0]["command"])
        html += f"<tr><td>{task['tags']['start_time']}</td><td style='background: #efefef;'>{command}</td><td>{task['tags']['duration']}</tr>\n"
    html += "</table>"
    html += "<hr>"
    html += svg + "\n"
    html += "<hr>\n"
    html += "</body>\n"
    html += "</html>\n"

    html_path = audit_path.replace(".au.json", ".au.html")
    with open(html_path, "w") as htmlfile:
        htmlfile.write(html)

    # Open html file in browser
    print(f"Trying to open HTML file in browser: {html_path} ...")
    sub.run(f"open {html_path}", shell=True)


def write_audit_files(
    command, input_paths, output_paths, start_time, end_time, merge_audit_files
):
    audit_extension = ".au.json"

    dur = end_time - start_time
    d = int(dur.days)
    h, rem = divmod(dur.seconds, 3600)
    m, rem = divmod(rem, 60)
    s = rem
    mus = int(dur.microseconds)

    s_float = float(f"{dur.seconds}.{dur.microseconds:06d}")

    iso_datetime_fmt = "%Y-%m-%dT%H:%M:%S.%fZ"

    inputs = [{"url": inpath, "path": None} for inpath in input_paths]
    outputs = [{"url": outpath, "path": None} for outpath in output_paths]

    audit_info = {
        "inputs": inputs,
        "outputs": outputs,
        "executors": {
            "image": None,
            "command": command.split(" "),
        },
        "tags": {
            "start_time": start_time.strftime(iso_datetime_fmt),
            "end_time": end_time.strftime(iso_datetime_fmt),
            "duration": f"{d}-{h:02d}:{m:02d}:{s:02d}.{mus:06d}",
            "duration_s": s_float,
        },
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
    pattern = re.compile("(\{i\:([^\{\}]+)\}|i\:([^\s\(\)]+))")
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
    pattern = re.compile("(\{o\:([^\{\}]+)\}|o\:([^\s\(\)]+))")
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
