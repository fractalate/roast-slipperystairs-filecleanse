package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var found int = 0
var path string
var extensions string
var recursive bool
var dryRun bool

func printMessage(remove bool, fileExtensions []string) {
	if !remove {
		message := fmt.Sprintf("filecleanse found %d total files using a %s extension.\n", found, extensions)
		if len(fileExtensions) > 1 {
			message = fmt.Sprintf("filecleanse found %d total files using file extensions %s.\n", found, extensions)
		}

		fmt.Print(message)
	}
}

func innerCheck(file string, remove bool, dir string) error {
	if !remove {
		found++
	} else if dryRun {
		fmt.Println(file)
	} else {
		err := os.Remove(dir)
		if err != nil {
			fmt.Println(err)
			return errors.New("File could not be removed: " + file)
		}
		fmt.Printf("Found and removed: %s\n", file)
	}

	return nil
}

func foundFile(fileExtension []string, s string) bool {
	return slices.ContainsFunc(fileExtension, func(ext string) bool {
		return strings.Contains(s, ext)
	})
}

func handleRecursive(fileExtensions []string, remove bool) error {
	err := filepath.WalkDir(path, func(recPath string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path: %s: %v\n", recPath, err)
			return err
		}

		isSecret := strings.HasPrefix(d.Name(), ".")
		isFile := d.IsDir()
		foundFile := foundFile(fileExtensions, d.Name())
		if !isSecret && isFile && foundFile {
			innerCheck(d.Name(), remove, recPath)
		}

		return nil
	})

	printMessage(remove, fileExtensions)
	return err
}

func handleFiles(files []os.DirEntry, remove bool, fileExtensions []string) error {
	for _, file := range files {
		// Skip hidden files
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		fileInfo, err := os.Stat(path + file.Name())
		if err != nil {
			fmt.Println(err)
			return errors.New("FileInfo could not be retrieved")
		}

		foundFile := foundFile(fileExtensions, fileInfo.Name())
		if !fileInfo.IsDir() && foundFile {
			innerCheck(fileInfo.Name(), remove, path+file.Name())
		}
	}

	printMessage(remove, fileExtensions)
	return nil
}

func validateResponse(response string) (bool, string) {
	trimmed := strings.ToLower(strings.TrimSpace(response))
	startsY := strings.HasPrefix(trimmed, "y")

	if !startsY ||
		len(trimmed) == 1 && trimmed != "y" ||
		len(trimmed) == 3 && trimmed != "yes" ||
		startsY && len(trimmed) > 3 {
		return false, "Input was either invalid or removal was cancelled!\n"
	}

	return true, ""
}

func checkFilePrefix(fileExtensions []string) []string {
	for _, ext := range fileExtensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
	}

	return fileExtensions
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

		files := []os.DirEntry{}
		fileExtensions := strings.Split(args[0], ",")
		extensions = strings.Join(fileExtensions, ", ")
		checkFilePrefix(fileExtensions)
		if len(path) == 0 {
			var err error
			path, err = os.Getwd()
			if err != nil {
				return errors.New("An error occurred when grabbing the current work directory")
			}
		}

		// Add to the trailing slash if needed
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		var err error
		response := ""
		if recursive {
			err = handleRecursive(fileExtensions, false)
		} else {
			files, err := os.ReadDir(path)
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
			fmt.Printf("Would you like to delete them? ")
			fmt.Scan(&response)
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
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "dr", false, "Provides a list of files that would be deleted without removing them")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to where your file extension exists")
}
