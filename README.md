# filecleanse
filecleanse is a CLI tool that will batch remove files that use a certain extension. By default it will use the current directory you are in, unless you specify your path using the `--path` subcommand. The tool will require the user to confirm they want to delete x number of files before removing them.

After cloning the repo, you can build the toll directly into `/usr/local/bin` (NOTE: `go build` needs to get executed in the project directory):

```
dcalligy at dcalligymacbook in ~/projects/filecleanse
$ sudo go build -o /usr/local/bin/filecleanse
```

**Basic Usage:**
```
dcalligy at dcalligymacbook in ~
$ filecleanse dat --path /Users/dcalligy/Downloads/
filecleanse found 5 total files using a .dat extension.
Would you like to delete them? y
Response:  y
Found and removed: 20260528(1).dat
Found and removed: 20260528(2).dat
Found and removed: 20260528(3).dat
Found and removed: 20260528(4).dat
Found and removed: 20260528.dat
```
