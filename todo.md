# Coding to do
 
- [x] Add basic tests
- [x] Check all substrings (separated by spaces) of current command if they
      correspond to an existing file.
- [=] If such a file exists and a corresponding audit file exists with the same
      command as the current one, identify it as an existing output file and
      skip running the current command. (i.e. skip the rest of the points
      below).
- [=] If the file instead lacks an audit file or has an audit file with a
      different command, identify it as an input file.
- [=] Create a temporary directory where the command can run.
- [=] Link all input files identified in 3) into the directory with symlinks
- [x] Run the command in the directory
- [x] Run a glob command in the directory to find all newly created files, and
      identify these as new output files
- [x] Create audit files for all of these (with the executed command and other
      info)
- [=] Move all output files to their final paths
- [=] Clear the temporary directory
