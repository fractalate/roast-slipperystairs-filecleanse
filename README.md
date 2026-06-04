# filecleanse
filecleanse is a CLI tool that will batch remove files that use a certain extension. By default it will use the current directory you are in, unless you specify your path using the `--path` subcommand. The tool will require the user to confirm they want to delete x number of files before removing them. Commas are used to delimit multiple file extensions

After cloning the repo, you can build the tool directly into `/usr/local/bin`.

```
// To build without changing directories, otherwise run sudo go build -o /usr/local/bin/filecleanse from the project directory
dilly@dilly:~$ (cd ~/projects/filecleanse && sudo go build -o /usr/local/bin/filecleanse .)
```

**Subcommands**
- `--path`: Used to specify the path of where the file extensions exist
- `--dry-run`: Will print out the potential files without deleting them
- `--recursive`: Find files recursively in the specified `path`. By default the CLI will only remove/find files at the top level of the specified `path`.

**Basic Usage:**
```
dilly@dilly:~$ filecleanse .dat --path /home/dilly/Downloads --recursive --dry-run
filecleanse found 3 total files using a .dat extension.
Would you like to delete them? yes
Potential file: /home/dilly/Downloads/mynested/file2.dat
Potential file: /home/dilly/Downloads/mynested/really/dir/gotcha.dat
Potential file: /home/dilly/Downloads/mynested/really/file3.dat

dilly@dilly:~$ filecleanse .dat, .txt --path /home/dilly/Downloads --recursive --dry-run
filecleanse found 4 total files using file extensions .dat, .txt.
Would you like to delete them? yes
Potential file: /home/dilly/Downloads/file1.txt
Potential file: /home/dilly/Downloads/mynested/file2.dat
Potential file: /home/dilly/Downloads/mynested/really/dir/gotcha.dat
Potential file: /home/dilly/Downloads/mynested/really/file3.dat

dilly@dilly:~$ filecleanse .dat, .xml, .txt, .js --path /home/dilly/Downloads/ --recursive
filecleanse found 6 total files using file extensions .dat, .xml, .txt, .js.
Would you like to delete them? yes
Removed file: /home/dilly/Downloads/index.js
Removed file: /home/dilly/Downloads/mynested/cant.dat
Removed file: /home/dilly/Downloads/mynested/really/deeply/main.xml
Removed file: /home/dilly/Downloads/mynested/really/deeply/nested/dir/lastone.txt
Removed file: /home/dilly/Downloads/mynested/really/deeply/nested/something.txt
Removed file: /home/dilly/Downloads/mynested/really/dis.dat
```
