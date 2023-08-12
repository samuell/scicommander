from pytest import fail
from os import path
import subprocess


def test_create_audit_file():
    outpath = ".tmp.hej.txt"
    #audit_path = f"{outpath}.audit.json"
    stdout, stderr, retcode = run_command(f"echo hej > {outpath}")
    print(
        f"""
    STDOUT: {stdout}
    STDERR: {stderr}
    RETCOD: {retcode}
    """
    )
    assert path.isfile(outpath)


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
