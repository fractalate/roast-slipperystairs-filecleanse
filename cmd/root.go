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

func printMessage(remove bool, fileExtensions []string) {
	if !remove {
		message := fmt.Sprintf("filecleanse found %d total files using a %s extension.\n", found, extensions)
		if len(fileExtensions) > 1 {
			message = fmt.Sprintf("filecleanse found %d total files using file extensions %s.\n", found, extensions)
		}

		fmt.Print(message)
	}
}

func handleRecursive(fileExtensions []string, remove bool) error {
	err := filepath.WalkDir(path, func(recPath string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path: %s: %v\n", recPath, err)
			return err
		}

		isSecret := !strings.HasPrefix(d.Name(), ".")
		isFile := d.IsDir()
		foundFile := slices.ContainsFunc(fileExtensions, func(ext string) bool {
			return strings.Contains(d.Name(), ext)
		})
		if !isSecret && isFile && foundFile {
			if !remove {
				found++
			} else {
				err := os.Remove(recPath)
				if err != nil {
					fmt.Println(err)
					return errors.New("Could no remove file: " + recPath)
				}

				fmt.Printf("Found and removed: %s\n", d.Name())
			}
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

		foundFile := slices.ContainsFunc(fileExtensions, func(ext string) bool {
			return strings.Contains(fileInfo.Name(), ext)
		})
		if !fileInfo.IsDir() && foundFile {
			if !remove {
				found++
			} else {
				err := os.Chdir(path)
				if err != nil {
					fmt.Println(err)
					return errors.New("Could not change directories to: " + path)
				}

				err = os.Remove(file.Name())
				if err != nil {
					fmt.Println(err)
					return errors.New("Could not remove file: " + file.Name())
				}
				fmt.Printf("Found and removed: %s\n", file.Name())
			}
		}
	}

	printMessage(remove, fileExtensions)
	return nil
}

func checkFilePrefix(fileExtensions []string) []string {
	for _, ext := range fileExtensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
	}

	return fileExtensions
}

// TODO => Add two more flags --recursive, and --dry-run
// TODO   - Recursive will use filepath.WalDir to traverse through the given path and remove the extensions provided.
// TODO   - Dry Run will print out a list of potential files to delete without removing them.
// TODO => Create a utility file and our root will call the necessary function based on the recursive flag.
// TODO => We could use another function that does the inner logic once we write out the new function.
// TODO => Cleanup some of the repetitiveness
var rootCmd = &cobra.Command{
	Use:   "filecleanse",
	Short: "Cleanse a directory from a specific file extension",
	Long: "filecleanse takes a file extension and a path to where those files exist\n" +
		"and removes them from your machine.\n\n" +
		"By default, filecleanse will look in the specified directory, report\n" +
		"anything.\n" +
		"Multiple file extensions should be comma delimited.",
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
			// os.Getwd() doesn't return the trailing slash in a directory.
			// Concatenate the trailing slash so we don't run into any errors.
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

		fmt.Println("Here we are: ", extensions)
		if found == 0 {
			fmt.Printf("Zero files were found using extension %s in path %s\n", extensions, path)
		} else {
			fmt.Printf("Would you like to delete them? ")
			fmt.Scan(&response)
			valid, msg := validateResponse(response)

			if !valid {
				fmt.Print(msg)
			} else if err = handleFiles(files, true, fileExtensions); err != nil && !recursive {
				return err
			} else if err = handleRecursive(fileExtensions, true); err != nil && recursive {
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
	rootCmd.Flags().BoolVar(&recursive, "recursive", false, "Recursively traverses the given path to delete files")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Provides a list of files that would be deleted without removing themd")
	rootCmd.Flags().StringVarP(&path, "path", "p", "", "Path to where your file extension exists")
}
