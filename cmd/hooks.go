package cmd

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var hooksCMD = &cobra.Command{
	Use:   "hooks",
	Short: "Install git hooks",
	Long:  "Manage hooks",
	PreRun: func(cmd *cobra.Command, args []string) {
		install, _ := cmd.Flags().GetBool("install")

		uninstall, _ := cmd.Flags().GetBool("uninstall")

		if install && uninstall {
			cmd.PrintErr("Cannot install and uninstall at the same time.")
			os.Exit(1)
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		install, _ := cmd.Flags().GetBool("install")
		uninstall, _ := cmd.Flags().GetBool("uninstall")

		const (
			HOOKSDIR    = "hooks"
			HOOKSYMLINK = ".git/hooks"
		)

		if install {

			files, err := os.ReadDir(HOOKSDIR)

			if err != nil {
				cmd.PrintErr(err.Error())
			}

			for _, file := range files {

				switch platform := runtime.GOOS; platform {
				case "linux":
					err := os.Symlink(filepath.Join(HOOKSDIR, file.Name()), filepath.Join(HOOKSYMLINK, file.Name()))
					if err != nil {
						cmd.PrintErr(err.Error())
					}
				case "windows":
					input, err := os.ReadFile(filepath.Join(HOOKSDIR, file.Name()))

					if err != nil {
						cmd.PrintErr(err.Error())
					}

					err = os.WriteFile(filepath.Join(HOOKSYMLINK, file.Name()), input, 0644)
					if err != nil {
						cmd.PrintErr(err.Error())
					}

					cmd.Printf("Installed %s\n", file.Name())

				default:
					cmd.PrintErr("Unsupported platform")

				}
			}

		} else if uninstall {

			files, err := os.ReadDir(HOOKSYMLINK)

			if err != nil {
				cmd.PrintErr(err.Error())
			}

			for _, file := range files {
				err := os.Remove(filepath.Join(HOOKSYMLINK, file.Name()))
				if err != nil {
					cmd.PrintErr(err.Error())
				}
				cmd.Printf("Uninstalled %s\n", file.Name())
			}

		} else {
			cmd.PrintErr("No action specified. Use --install or --uninstall.")
		}
	},
}

func init() {
	rootCmd.AddCommand(hooksCMD)

	hooksCMD.Flags().Bool("install", false, "Install hooks")
	hooksCMD.Flags().Bool("uninstall", false, "Uninstall hooks")
}
