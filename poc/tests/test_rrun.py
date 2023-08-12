from pytest import fail
import os
import subprocess


def test_create_file():
    tmpdir = ".tmp.scicmd-test"
    os.makedirs(tmpdir, exist_ok=True)
    cmd = f"python scicmd.py -c 'echo hej > o:{tmpdir}/hej.txt'"
    run_command(cmd)
    assert os.path.isfile(f"{tmpdir}/hej.txt")
    os.remove(f"{tmpdir}/hej.txt")
    os.remove(f"{tmpdir}/hej.txt.au.json")
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
