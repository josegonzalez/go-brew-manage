package main

import (
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type BrewYaml []map[string]interface{}

func main() {
	configFile := flag.String("config", "brew.yaml", "path to the brew.yaml config file")
	flag.Parse()

	fmt.Printf("reading %v\n", *configFile)
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Printf("Could not open YAML file: %s", err.Error())
		return
	}

	fmt.Printf("parsing %v\n", *configFile)
	var settings BrewYaml
	err = yaml.Unmarshal(data, &settings)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	var casks BrewYaml
	var formulae BrewYaml
	var gemPackages BrewYaml
	var pipPackages BrewYaml
	var taps BrewYaml
	for _, entry := range settings {
		if _, ok := entry["homebrew_cask"]; ok {
			casks = append(casks, entry)
		} else if _, ok := entry["homebrew_formula"]; ok {
			formulae = append(formulae, entry)
		} else if _, ok := entry["homebrew_gem"]; ok {
			gemPackages = append(gemPackages, entry)
		} else if _, ok := entry["homebrew_pip"]; ok {
			pipPackages = append(pipPackages, entry)
		} else if _, ok := entry["homebrew_tap"]; ok {
			taps = append(taps, entry)
		}
	}

	if len(casks) > 0 {
		taps = appendCaskTaps(taps)
	}

	if len(pipPackages) > 0 {
		formulae = appendFormula(formulae, []string{"python", "brew-pip"})
	}

	if len(gemPackages) > 0 {
		formulae = appendFormula(formulae, []string{"brew-gem"})
	}

	installedModifier := func(l []string) []string { return l }

	brewUpdate()
	manageBrewCollection(taps, "task", []string{"tap", "--quieter"}, []string{"tap", "--quieter"}, installedModifier)
	brewUpdate()

	manageBrewCollection(casks, "cask", []string{"cask", "list"}, []string{"cask", "install"}, installedModifier)
	manageBrewCollection(formulae, "formula", []string{"list"}, []string{"install"}, installedModifier)

	pipListArguments := []string{"pip", "list"}
	pipInstallArguments := []string{"pip", "install"}
	installedPipModifier := func(l []string) []string { return l }
	if len(pipPackages) > 0 {
		cmd := exec.Command("brew", "pip", "--version")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HOMEBREW_NO_AUTO_UPDATE=1")
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Printf("pip: state=error %v\n", err)
			return
		}

		if strings.HasPrefix(string(stdout), "brew pip v0.4.") {
			pipListArguments = []string{"list"}
			pipInstallArguments = []string{"pip"}
			installedPipModifier = func(l []string) []string {
				var installedPipPackages []string
				for _, formula := range l {
					if strings.HasPrefix(formula, "pip-") {
						installedPipPackages = append(installedPipPackages, strings.TrimPrefix(formula, "pip-"))
					}
				}
				return installedPipPackages
			}
		}
	}

	manageBrewCollection(pipPackages, "pip", pipListArguments, pipInstallArguments, installedPipModifier)

	gemListArguments := []string{"list"}
	gemInstallArguments := []string{"gem", "install"}
	installedGemModifier := func(l []string) []string {
		var installedGemPackages []string
		for _, formula := range l {
			if strings.HasPrefix(formula, "gem-") {
				installedGemPackages = append(installedGemPackages, strings.TrimPrefix(formula, "gem-"))
			}
		}
		return installedGemPackages
	}

	manageBrewCollection(gemPackages, "gem", gemListArguments, gemInstallArguments, installedGemModifier)
}

func brewUpdate() {
	fmt.Println("brew: updating")
	stdout, err := exec.Command("brew", "update").CombinedOutput()
	if err != nil {
		fmt.Printf("brew: state=error %v\n", stdout)
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func manageBrewCollection(entries BrewYaml, entryType string, listArguments []string, installArguments []string, installedModifier func(l []string) []string) bool {
	if len(entries) == 0 {
		return true
	}

	fmt.Printf("%s: fetching\n", entryType)
	cmd := exec.Command("brew", listArguments...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HOMEBREW_NO_AUTO_UPDATE=1")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s: state=error %v\n", entryType, err)
		return false
	}
	installedEntries := strings.Split(string(stdout), "\n")
	installedEntries = installedModifier(installedEntries)

	hasErrors := false
	for _, entry := range entries {
		value, ok := entry["name"]
		if !ok {
			fmt.Printf("%s: state=name-error %v\n", entryType, entry)
			hasErrors = true
			continue
		}

		name := value.(string)
		fmt.Printf("%s: name=%v", entryType, name)
		if stringInSlice(name, installedEntries) {
			fmt.Printf(" state=present\n")
			continue
		}

		installArguments = append(installArguments, name)
		cmd := exec.Command("brew", installArguments...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HOMEBREW_NO_AUTO_UPDATE=1")
		installOutput, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf(" state=install-error %v\n", string(installOutput))
			hasErrors = true
			continue
		}
		fmt.Printf(" state=installed\n")
	}
	return hasErrors
}

func appendCaskTaps(taps BrewYaml) BrewYaml {
	caskTapNames := []string{"homebrew/cask", "homebrew/cask-drivers", "homebrew/cask-fonts", "homebrew/cask-versions"}
	var tapNames []string
	for _, entry := range taps {
		value, ok := entry["name"]
		if !ok {
			continue
		}
		tapNames = append(tapNames, value.(string))
	}

	for _, tapName := range caskTapNames {
		if !stringInSlice(tapName, tapNames) {
			tap := make(map[string]interface{})
			tap["homebrew_cask"] = nil
			tap["name"] = tapName
			taps = append(taps, tap)
		}
	}

	return taps
}

func appendFormula(formulae BrewYaml, formulaeToAppend []string) BrewYaml {
	var formulaNames []string
	for _, entry := range formulae {
		value, ok := entry["name"]
		if !ok {
			continue
		}
		formulaNames = append(formulaNames, value.(string))
	}

	for _, formulaName := range formulaeToAppend {
		if !stringInSlice(formulaName, formulaNames) {
			formula := make(map[string]interface{})
			formula["homebrew_formula"] = nil
			formula["name"] = formulaName
			formulae = append(formulae, formula)
		}
	}

	return formulae
}
