import os
import sys

try:
    from setuptools import setup
except:
    from distutils.core import setup

with open("../README.md") as fobj:
    long_description = fobj.read()

setup(
    name="scicommander",
    version="0.3.2",
    description="A small library for executing shell commands in a reproducible way.",
    long_description=long_description,
    long_description_content_type="text/markdown",
    author="Samuel Lampa",
    author_email="samuel.lampa@scilifelab.se",
    url="https://github.com/samuell/scicommander",
    license="MIT",
    keywords=["reproducibility", "shell", "bash"],
    packages=[
        "scicommander",
    ],
    classifiers=[
        "Development Status :: 4 - Beta",
        "Environment :: Console",
        "Intended Audience :: Science/Research",
        "License :: OSI Approved :: MIT License",
        "Natural Language :: English",
        "Operating System :: POSIX :: Linux",
        "Programming Language :: Python",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.7",
        "Topic :: Scientific/Engineering",
        "Topic :: Scientific/Engineering :: Bio-Informatics",
    ],
    entry_points={
        "console_scripts": [
            "scicmd=scicommander.scicmd:main",
        ],
    },
    scripts=["scripts/scishell"]
)
