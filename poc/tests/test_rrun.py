from pytest import fail
import json
import os
import subprocess


def test_create_file():
    tmpdir = ".tmp.scicmd-test-create-file"
    os.makedirs(tmpdir, exist_ok=True)
    cmd = f"python scicmd.py -c 'echo hej > o:{tmpdir}/hej.txt'"
    run_command(cmd)

    # Make sure everything is as it should
    assert os.path.isfile(f"{tmpdir}/hej.txt")

    # Clean up
    os.remove(f"{tmpdir}/hej.txt")
    os.remove(f"{tmpdir}/hej.txt.au.json")
    os.removedirs(tmpdir)


def test_create_two_files():
    tmpdir = ".tmp.scicmd-test-create-two-files"
    os.makedirs(tmpdir, exist_ok=True)
    cmd1 = f"python scicmd.py -c 'echo hej > o:{tmpdir}/hej.txt'"
    run_command(cmd1)
    cmd2 = f"python scicmd.py -c 'echo $(cat i:{tmpdir}/hej.txt) da > o:{tmpdir}/hej.da.txt'"
    run_command(cmd2)

    # Make sure everything is as it should
    assert os.path.isfile(f"{tmpdir}/hej.txt")
    assert os.path.isfile(f"{tmpdir}/hej.da.txt")

    with open(f"{tmpdir}/hej.txt") as outfile1:
        content = outfile1.read().strip()
        assert content == "hej"

    with open(f"{tmpdir}/hej.da.txt") as outfile2:
        content = outfile2.read().strip()
        assert content == "hej da"

    with open(f"{tmpdir}/hej.txt.au.json") as aufile1:
        audit_info = json.load(aufile1)
        want_dict1 = {
            "command": f"echo hej > {tmpdir}/hej.txt",
            "inputs": [],
            "outputs": [f"{tmpdir}/hej.txt"],
            "upstream": {},
        }
        assert audit_info == want_dict1

    with open(f"{tmpdir}/hej.da.txt.au.json") as aufile2:
        audit_info2 = json.load(aufile2)
        want_dict2 = {
            "command": f"echo $(cat {tmpdir}/hej.txt) da > {tmpdir}/hej.da.txt",
            "inputs": [f"{tmpdir}/hej.txt"],
            "outputs": [f"{tmpdir}/hej.da.txt"],
            "upstream": {},
        }
        assert audit_info2 == want_dict2

    # Clean up
    os.remove(f"{tmpdir}/hej.txt")
    os.remove(f"{tmpdir}/hej.txt.au.json")
    os.remove(f"{tmpdir}/hej.da.txt")
    os.remove(f"{tmpdir}/hej.da.txt.au.json")
    os.removedirs(tmpdir)


def run_command(command):
    try:
        # Run the command and capture stdout, stderr, and return code
        result = subprocess.run(
            command,
            shell=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )

        # Get the stdout, stderr, and return code
        stdout = result.stdout.strip()
        stderr = result.stderr.strip()
        return_code = result.returncode

        return stdout, stderr, return_code

    except subprocess.CalledProcessError as e:
        # If there's an error with the command, handle it here if needed
        print(f"Error occurred: {e}")
        return "", str(e), e.returncode
