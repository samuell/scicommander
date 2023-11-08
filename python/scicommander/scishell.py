from cmd import Cmd
from subprocess import Popen, PIPE
import os

from scicommander import scicmd


def main():
    scishell = SciShell()
    scishell.prompt = """\033[32mscicmdr > \033[39m"""
    scishell.cmdloop()


class SciShell(Cmd):
    def default(self, input_text):
        if input_text in ["q", "exit"]:
            return True

        try:
            scicmd.parse_and_execute(input_text)
        except Exception as e:
            print(f"ERROR: {e}")

    def do_EOF(self, _):
        print("Exiting SciCommander ...")
        return True


if __name__ == "__main__":
    main()
