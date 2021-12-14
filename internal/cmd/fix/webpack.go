package fix

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/gobuffalo/cli/internal/genny/assets/webpack"
)

// WebpackCheck will compare the current default Buffalo
// webpack.config.js against the applications webpack.config.js. If they are
// different you have the option to overwrite the existing webpack.config.js
// file with the new one.
func WebpackCheck(r *Runner) error {
	fmt.Println("~~~ Checking webpack.config.js ~~~")

	if !r.App.WithWebpack {
		return nil
	}

	templates, err := webpack.Templates()
	if err != nil {
		return err
	}

	tmpl, err := template.New("webpack.config.js.tmpl").ParseFS(templates, "webpack.config.js.tmpl")
	if err != nil {
		return err
	}

	bb := &bytes.Buffer{}
	err = tmpl.Execute(bb, map[string]interface{}{
		"opts": &webpack.Options{
			App: r.App,
		},
	})
	if err != nil {
		return err
	}

	b, err := os.ReadFile("webpack.config.js")
	if err != nil {
		return err
	}

	if string(b) == bb.String() {
		return nil
	}

	if !ask("Your webpack.config.js file is different from the latest Buffalo template.\nWould you like to replace yours with the latest template?") {
		fmt.Println("\tSkipping webpack.config.js")
		return nil
	}

	wf, err := os.Create("webpack.config.js")
	if err != nil {
		return err
	}
	_, err = wf.Write(bb.Bytes())
	if err != nil {
		return err
	}
	return wf.Close()
}
