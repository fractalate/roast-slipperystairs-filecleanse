package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var found int = 0
var path string
var extensions string
var recursive bool
var dryRun bool
var enlightenment bool

func printMessage(remove bool, fileExtensions []string) {
	if !remove {
		message := fmt.Sprintf("filecleanse found %d total files using a %s extension.\n", found, extensions)
		if len(fileExtensions) > 1 {
			message = fmt.Sprintf("filecleanse found %d total files using file extensions %s.\n", found, extensions)
		}

		fmt.Print(message)
	}
}

func removeFile(remove bool, dir string) error {
	if !remove {
		found++
	} else if dryRun {
		fmt.Println("Potential file:", dir)
	} else {
		err := os.Remove(dir)
		if err != nil {
			return err
		}
		fmt.Printf("Removed file: %s\n", dir)
	}

	return nil
}

func findFile(fileExtension []string, s string) bool {
	return slices.ContainsFunc(fileExtension, func(ext string) bool {
		return filepath.Ext(s) == ext
	})
}

func handleRecursive(fileExtensions []string, remove bool) error {
	err := filepath.WalkDir(path, func(recPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		isSecret := strings.HasPrefix(d.Name(), ".")
		isNotAFile := d.IsDir()
		found := findFile(fileExtensions, d.Name())
		if !isSecret && !isNotAFile && found {
			err = removeFile(remove, recPath)
			if err != nil {
				return err
			}
		}

		return nil
	})

	printMessage(remove, fileExtensions)
	return err
}

func handleFiles(files []os.DirEntry, remove bool, fileExtensions []string) error {
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		fileInfo, err := os.Stat(path + file.Name())
		if err != nil {
			return err
		}

		found := findFile(fileExtensions, fileInfo.Name())
		if !fileInfo.IsDir() && found {
			err = removeFile(remove, path+fileInfo.Name())
			if err != nil {
				return err
			}
		}
	}

	printMessage(remove, fileExtensions)
	return nil
}

func validateResponse(response string) (bool, string) {
	trimmed := strings.ToLower(strings.TrimSpace(response))
	if len(trimmed) == 0 {
		return false, "Response cannot be empty\n"
	}

	startsY := strings.HasPrefix(trimmed, "y")
	if !startsY ||
		len(trimmed) == 1 && trimmed != "y" ||
		len(trimmed) == 2 && trimmed != "ye" ||
		len(trimmed) == 3 && trimmed != "yes" ||
		startsY && len(trimmed) > 3 {
		return false, "Input was either invalid or removal was cancelled!\n"
	}

	return true, ""
}

func validateArgs(args []string) ([]string, error) {
	extensions := []string{}
	var validExtPattern = regexp.MustCompile(`^\.?[a-zA-Z0-9]+$`)
	for i := range args {
		// Add file prefix if it doesn't exist
		ext := strings.ReplaceAll(args[i], ",", "")
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}

		if !validExtPattern.MatchString(ext) {
			extensions = []string{}
			msg := "Invalid extension supplied to filecleanse: " + ext
			return extensions, errors.New(msg)
		}
		extensions = append(extensions, ext)
	}
	return extensions, nil
}

var rootCmd = &cobra.Command{
	Use:   "filecleanse",
	Short: "Cleanse a directory from a specific file extension",
	Long: "filecleanse removes all the files matching a given extension from a specified path.\n" +
		"By default, filecleanse will look in your current working directory," +
		" unless specified with the --path subcommand.\nMultiple file extensions must be comma delimited, otherwise it might blow up.\n",
	Example: "  Remove all .log files from a specific directory\n" +
		"  filecleanse log --path /var/log\n",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Missing arguments! Type filecleanse --help for CLI usage.")
		}

		if enlightenment {
			p := path
			if len(p) == 0 {
				p = "."
			}
			exts, _ := validateArgs(args)
			nameArgs := []string{}
			for _, ext := range exts {
				nameArgs = append(nameArgs, fmt.Sprintf("-name \"*%s\"", ext))
			}
			nameExpr := strings.Join(nameArgs, " -o ")
			if len(exts) > 1 {
				nameExpr = "\\( " + nameExpr + " \\)"
			}
			depth := ""
			if !recursive {
				depth = "-maxdepth 1 "
			}
			action := "-delete"
			if dryRun {
				action = "-print"
			}
			fmt.Printf("You never needed filecleanse. You needed:\n\n")
			fmt.Printf("  find %s %s%s %s\n\n", p, depth, nameExpr, action)
			fmt.Printf("This has worked since 1995. You're welcome.\n")
			return nil
		}

		var err error
		files := []os.DirEntry{}
		fileExtensions, err := validateArgs(args)
		if err != nil {
			return err
		}

		extensions = strings.Join(fileExtensions, ", ")
		if len(path) == 0 {
			path, err = os.Getwd()
			if err != nil {
				return errors.New("An error occurred when grabbing the current work directory")
			}
		}

		// Add to the trailing slash if needed
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		if recursive {
			err = handleRecursive(fileExtensions, false)
		} else {
			files, err = os.ReadDir(path)
			if err != nil {
				return errors.New("Directory: " + path + " could not be read")
			}

			err = handleFiles(files, false, fileExtensions)
			if err != nil {
				return err
			}
		}

		if found == 0 {
			fmt.Printf("Zero files were found using extension %s in path %s\n", extensions, path)
		} else {
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Printf("Would you like to delete them? ")
			if scanner.Scan() {
				response := scanner.Text()
				valid, msg := validateResponse(response)

				if !valid {
					fmt.Print(msg)
				} else {
					if recursive {
						err = handleRecursive(fileExtensions, true)
					} else {
						err = handleFiles(files, true, fileExtensions)
					}

					if err != nil {
						return err
					}
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively traverses the given path to delete files")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Provides a list of files that would be deleted without removing them")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to where your file extension exists")
	rootCmd.Flags().BoolVarP(&enlightenment, "enlightenment", "e", false, "Shows you the shell command that makes this program unnecessary")
}
