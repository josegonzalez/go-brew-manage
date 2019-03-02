package main

import (
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
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
	var pipPackages BrewYaml
	var taps BrewYaml
	for _, entry := range settings {
		if _, ok := entry["homebrew_cask"]; ok {
			casks = append(casks, entry)
		} else if _, ok := entry["homebrew_formula"]; ok {
			formulae = append(formulae, entry)
		} else if _, ok := entry["homebrew_pip"]; ok {
			pipPackages = append(pipPackages, entry)
		} else if _, ok := entry["homebrew_tap"]; ok {
			taps = append(taps, entry)
		}
	}

	installTaps(taps, casks)
	installCasks(casks)
	installFormulae(formulae, pipPackages)
	installPipPackages(pipPackages)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func installTaps(taps BrewYaml, casks BrewYaml) bool {
	fmt.Println("taps: fetching")
	stdout, err := exec.Command("brew", "tap", "--quieter").Output()
	if err != nil {
		fmt.Printf("tap: state=error %v\n", err)
		return false
	}
	installedTaps := strings.Split(string(stdout), "\n")

	if len(casks) > 0 {
		taps = appendCaskTaps(taps)
	}

	hasErrors := false
	for _, entry := range taps {
		value, ok := entry["name"]
		if !ok {
			fmt.Printf("tap: state=name-error %v\n", entry)
			hasErrors = true
			continue
		}

		name := value.(string)
		fmt.Printf("tap: name=%v", name)
		if stringInSlice(name, installedTaps) {
			fmt.Printf(" state=present\n")
			continue
		}
		installOutput, err := exec.Command("brew", "tap", "--quieter", name).Output()
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

func installCasks(casks BrewYaml) bool {
	fmt.Println("cask: fetching")
	stdout, err := exec.Command("brew", "cask", "list").Output()
	if err != nil {
		fmt.Printf("cask: state=error %v\n", err)
		return false
	}
	installedCasks := strings.Split(string(stdout), "\n")

	hasErrors := false
	for _, entry := range casks {
		value, ok := entry["name"]
		if !ok {
			fmt.Printf("cask: state=name-error %v\n", entry)
			hasErrors = true
			continue
		}

		name := value.(string)
		fmt.Printf("cask: name=%v", name)
		if stringInSlice(name, installedCasks) {
			fmt.Printf(" state=present\n")
			continue
		}
		installOutput, err := exec.Command("brew", "cask", "install", name).Output()
		if err != nil {
			fmt.Printf(" state=install-error %v\n", string(installOutput))
			hasErrors = true
			continue
		}
		fmt.Printf(" state=installed\n")
	}

	return hasErrors
}

func installFormulae(formulae BrewYaml, pipPackages BrewYaml) bool {
	fmt.Println("formula: fetching")
	stdout, err := exec.Command("brew", "list").Output()
	if err != nil {
		fmt.Printf("formula: state=error %v\n", err)
		return false
	}
	installedCasks := strings.Split(string(stdout), "\n")

	if len(pipPackages) > 0 {
		formulae = appendPythonFormula(formulae)
	}

	hasErrors := false
	for _, entry := range formulae {
		value, ok := entry["name"]
		if !ok {
			fmt.Printf("formula: state=name-error %v\n", entry)
			hasErrors = true
			continue
		}

		name := value.(string)
		fmt.Printf("formula: name=%v", name)
		if stringInSlice(name, installedCasks) {
			fmt.Printf(" state=present\n")
			continue
		}
		installOutput, err := exec.Command("brew", "install", name).Output()
		if err != nil {
			fmt.Printf(" state=install-error %v\n", string(installOutput))
			hasErrors = true
			continue
		}
		fmt.Printf(" state=installed\n")
	}

	return hasErrors
}

func appendPythonFormula(formulae BrewYaml) BrewYaml {
	addPython := true
	addBrewPip := true
	for _, entry := range formulae {
		value, ok := entry["name"]
		if !ok {
			continue
		}

		if value.(string) == "python" {
			addPython = false
			continue
		}

		if value.(string) == "brew-pip" {
			addBrewPip = false
			continue
		}
	}

	if addPython {
		tap := make(map[string]interface{})
		tap["homebrew_formula"] = nil
		tap["name"] = "python"
		formulae = append(formulae, tap)
	}

	if addBrewPip {
		tap := make(map[string]interface{})
		tap["homebrew_formula"] = nil
		tap["name"] = "brew-pip"
		formulae = append(formulae, tap)
	}
	return formulae
}

func installPipPackages(pipPackages BrewYaml) bool {
	fmt.Println("pip: fetching")
	stdout, err := exec.Command("brew", "list").Output()
	if err != nil {
		fmt.Printf("tap: state=error %v\n", err)
		return false
	}

	installedFormula := strings.Split(string(stdout), "\n")
	var installedPipPackages []string
	for _, formula := range installedFormula {
		if strings.HasPrefix(formula, "pip-") {
			installedPipPackages = append(installedPipPackages, formula)
		}
	}

	hasErrors := false
	for _, entry := range pipPackages {
		value, ok := entry["name"]
		if !ok {
			fmt.Printf("pip: state=name-error %v\n", entry)
			hasErrors = true
			continue
		}

		name := value.(string)
		fmt.Printf("pip: name=%v", name)
		if stringInSlice(fmt.Sprintf("pip-%s", name), installedPipPackages) {
			fmt.Printf(" state=present\n")
			continue
		}
		installOutput, err := exec.Command("brew", "pip", name).Output()
		if err != nil {
			fmt.Printf(" state=install-error %v\n", string(installOutput))
			hasErrors = true
			continue
		}
		fmt.Printf(" state=installed\n")
	}

	return hasErrors
}
