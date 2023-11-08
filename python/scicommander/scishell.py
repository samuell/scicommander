from cmd import Cmd
from subprocess import Popen, PIPE
import os


def main():
    scishell = SciShell()
    scishell.prompt = """\033[32mscicmdr > \033[39m"""
    scishell.cmdloop()


class SciShell(Cmd):
    def default(self, input_text):
        if input_text in ["q", "exit"]:
            return True
        p = Popen(input_text, shell=True, stdin=PIPE, stdout=PIPE, stderr=PIPE)
        output, err = p.communicate()
        rc = p.returncode
        outstr = bytes.decode(output).strip()
        print(outstr)

    def do_EOF(self, _):
        print("Exiting SciCommander ...")
        return True


if __name__ == "__main__":
    main()
