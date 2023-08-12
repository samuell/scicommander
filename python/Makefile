#!/bin/bash
build: setup.py
	python setup.py sdist bdist_wheel

upload: dist
	twine upload dist/*
