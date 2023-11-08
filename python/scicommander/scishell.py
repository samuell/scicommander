from cmd import Cmd


def main():
    scishell = SciShell()
    scishell.prompt = """\033[32mscicmdr > \033[39m"""
    scishell.cmdloop()


class SciShell(Cmd):
    pass


if __name__ == "__main__":
    main()
