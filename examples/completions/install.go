package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/starriver/charli"
)

var installDescription = `
Installs bash and fish completions.
`

var install = charli.Command{
	Name:        "install",
	Headline:    "Install completions",
	Description: installDescription,
	Run: func(r *charli.Result) bool {
		if r.Fail {
			return false
		}

		bashInstalled, err := InstallBash()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false
		}

		fishInstalled, err := InstallFish()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false
		}

		if !(bashInstalled || fishInstalled) {
			fmt.Println("Nothing to install.")
			return false
		}

		return true
	},
}

func InstallBash() (bool, error) {
	if _, err := exec.LookPath("bash"); err != nil {
		fmt.Println("bash not installed")
		return false, nil
	}

	dir := filepath.Join(xdg.DataHome, "bash_completion", "completions")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, err
	}

	path := filepath.Join(dir, "charli-example")
	f, err := os.Create(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	charli.GenerateBashCompletions(f, "completions", "--_complete")

	fmt.Printf("bash completions installed to: %s\n", path)
	return true, nil
}

func InstallFish() (bool, error) {
	if _, err := exec.LookPath("fish"); err != nil {
		fmt.Println("fish not installed")
		return false, nil
	}

	dir := filepath.Join(xdg.DataHome, "fish", "vendor_completions.d")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, err
	}

	path := filepath.Join(dir, "charli-example")
	f, err := os.Create(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	charli.GenerateFishCompletions(f, "completions", "--_complete")

	fmt.Printf("fish completions installed to: %s\n", path)
	return true, nil
}
