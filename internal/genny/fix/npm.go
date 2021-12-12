package fix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gobuffalo/cli/internal/genny/assets/webpack"
)

// AddPackageJSONScripts rewrites the package.json file
// to add dev and build scripts if there are missing.
func AddPackageJSONScripts(opts *Options) ([]string, error) {
	if !opts.App.WithWebpack {
		return nil, nil
	}
	fmt.Println("~~~ Patching package.json to add dev and build scripts ~~~")

	b, err := os.ReadFile("package.json")
	if err != nil {
		return nil, err
	}

	needRewrite := false
	packageJSON := map[string]interface{}{}
	if err := json.Unmarshal(b, &packageJSON); err != nil {
		return nil, fmt.Errorf("could not rewrite package.json: %s", err.Error())
	}

	if _, ok := packageJSON["scripts"]; !ok {
		needRewrite = true
		// Add scripts
		packageJSON["scripts"] = map[string]string{
			"dev":   "webpack --watch",
			"build": "webpack -p --progress",
		}
	} else {
		// Add missing scripts
		scripts, ok := packageJSON["scripts"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("could not rewrite package.json: invalid scripts section")
		}
		if _, ok := scripts["dev"]; !ok {
			needRewrite = true
			scripts["dev"] = "webpack --watch"
		}
		if _, ok := scripts["build"]; !ok {
			needRewrite = true
			scripts["build"] = "webpack -p --progress"
		}
		packageJSON["scripts"] = scripts
	}

	if needRewrite {
		b, err = json.MarshalIndent(packageJSON, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("could not rewrite package.json: %w", err)
		}

		if err := os.WriteFile("package.json", b, 0o644); err != nil {
			return nil, fmt.Errorf("could not rewrite package.json: %w", err)
		}
	} else {
		fmt.Println("~~~ package.json doesn't need to be patched, skipping. ~~~")
	}

	return nil, nil
}

// PackageJSONCheck will compare the current default Buffalo
// package.json against the applications package.json. If they are
// different you have the option to overwrite the existing package.json
// file with the new one.
func PackageJSONCheck(opts *Options) ([]string, error) {
	fmt.Println("~~~ Checking package.json ~~~")

	if !opts.App.WithWebpack {
		return nil, nil
	}

	templates, err := webpack.Templates()
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("package.json").ParseFS(templates, "package.json.tmpl")
	if err != nil {
		return nil, err
	}

	bb := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(bb, "package.json.tmpl", map[string]interface{}{
		"opts": &webpack.Options{
			App: opts.App,
		},
	})
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile("package.json")
	if err != nil {
		return nil, err
	}

	if string(b) == bb.String() {
		return nil, nil
	}

	if !opts.YesToAll && !ask("Your package.json file is different from the latest Buffalo template.\nWould you like to REPLACE yours with the latest template?") {
		fmt.Println("\tskipping package.json")
		return nil, nil
	}

	pf, err := os.Create("package.json")
	if err != nil {
		return nil, err
	}
	_, err = pf.Write(bb.Bytes())
	if err != nil {
		return nil, err
	}
	err = pf.Close()
	if err != nil {
		return nil, err
	}

	os.RemoveAll(filepath.Join(opts.App.Root, "node_modules"))
	var cmd *exec.Cmd
	if opts.App.WithYarn {
		cmd = exec.Command("yarnpkg", "install")
	} else {
		cmd = exec.Command("npm", "install")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return nil, cmd.Run()
}
