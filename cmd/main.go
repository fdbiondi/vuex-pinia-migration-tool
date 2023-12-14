package main

import (
	"fileutil"
	"fmt"
	"os"
	"parser"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	verbose    bool
	debug      bool
	removeDest bool
)

var rootCmd = &cobra.Command{
	Use:   "vuex-2-pinia",
	Short: "A migration tool for vuex code base to pinia state management format",
}

func main() {

	migrateCmd := &cobra.Command{
		Use:   "migrate [SOURCE PATH] [DEST PATH]",
		Short: "Translates code from a source directory written in vuex to a destiny directory",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			// set flags
			parser.Verbose = verbose
			parser.Debug = debug

			// grab directories
			sourceDir, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			destDir, err := filepath.Abs(args[1])
			if err != nil {
				return err
			}

			if !fileutil.Exists(sourceDir) {
				return fmt.Errorf("source directory '%s' does not exist", sourceDir)
			}

			if removeDest {
				err := os.RemoveAll(destDir)
				if err != nil {
					return err
				}
			} else {
				if fileutil.Exists(destDir) {
					// move old dir
					err = fileutil.VersionDir(destDir)
					if err != nil {
						return err
					}

					// create new dir
					err = fileutil.CreateIfNotExists(destDir, 0755)
					if err != nil {
						return err
					}
				}
			}

			err = fileutil.CopyDirectory(sourceDir, destDir)
			if err != nil {
				return err
			}

			if parser.Verbose {
				fmt.Printf("source path '%s'\n", sourceDir)
				fmt.Printf("output path '%s'\n\n", destDir)
			}

			err = parser.Execute(destDir)
			if err != nil {
				return err
			} else {
				fmt.Println("migrated!")
			}

			return nil
		},
	}

	migrateCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	migrateCmd.PersistentFlags().BoolVarP(&removeDest, "remove-dest", "r", false, "remove destiny directory")
	migrateCmd.PersistentFlags().BoolVarP(&debug, "debug-mode", "d", false, "enable debug mode")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Vuex2Pinia",
		Long:  `All software has versions. This is Vuex2Pinia's`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("Vuex2Pinia migrate tool v0.1")
		},
	}

	rootCmd.SetVersionTemplate("Vuex2Pinia migrate tool v0.1")

	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(versionCmd)

	rootCmd.Execute()
}
