package fix

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

var replace = map[string]string{
	"github.com/gobuffalo/buffalo-plugins":         "github.com/gobuffalo/cli/internal/plugins",
	"github.com/gobuffalo/genny":                   "github.com/gobuffalo/genny/v2",
	"github.com/gobuffalo/pop":                     "github.com/gobuffalo/pop/v6",
	"github.com/gobuffalo/pop/nulls":               "github.com/gobuffalo/nulls",
	"github.com/gobuffalo/uuid":                    "github.com/gofrs/uuid",
	"github.com/markbates/pop":                     "github.com/gobuffalo/pop/v5",
	"github.com/markbates/validate":                "github.com/gobuffalo/validate/v3",
	"github.com/markbates/willie":                  "github.com/gobuffalo/httptest",
	"github.com/satori/go.uuid":                    "github.com/gofrs/uuid",
	"github.com/shurcooL/github_flavored_markdown": "github.com/gobuffalo/github_flavored_markdown",
	"github.com/gobuffalo/validate":                "github.com/gobuffalo/validate/v3",
	"github.com/gobuffalo/validate/validators":     "github.com/gobuffalo/validate/v3/validators",
	"github.com/gobuffalo/suite":                   "github.com/gobuffalo/suite/v4",
	"github.com/gobuffalo/buffalo-pop/":            "github.com/gobuffalo/buffalo-pop/v3",
	"github.com/gobuffalo/buffalo-pop/pop/popmw":   "github.com/gobuffalo/buffalo-pop/v3/pop/popmw",
	"github.com/gobuffalo/plush":                   "github.com/gobuffalo/plush/v4",
	"github.com/gobuffalo/mw-i18n":                 "github.com/gobuffalo/mw-i18n/v2",
}

var ic = ImportConverter{
	Data: replace,
}

var mr = MiddlewareTransformer{
	PackagesReplacement: map[string]string{
		"github.com/gobuffalo/buffalo/middleware/basicauth": "github.com/gobuffalo/mw-basicauth",
		"github.com/gobuffalo/buffalo/middleware/csrf":      "github.com/gobuffalo/mw-csrf",
		"github.com/gobuffalo/buffalo/middleware/i18n":      "github.com/gobuffalo/mw-i18n",
		"github.com/gobuffalo/buffalo/middleware/ssl":       "github.com/gobuffalo/mw-forcessl",
		"github.com/gobuffalo/buffalo/middleware/tokenauth": "github.com/gobuffalo/mw-tokenauth",
	},

	Aliases: map[string]string{
		"github.com/gobuffalo/mw-basicauth":   "basicauth",
		"github.com/gobuffalo/mw-contenttype": "contenttype",
		"github.com/gobuffalo/mw-csrf":        "csrf",
		"github.com/gobuffalo/mw-forcessl":    "forcessl",
		"github.com/gobuffalo/mw-i18n":        "i18n",
		"github.com/gobuffalo/mw-paramlogger": "paramlogger",
		"github.com/gobuffalo/mw-tokenauth":   "tokenauth",
	},
}

var checks = []Check{
	ic.Process,
	mr.transformPackages,
	WebpackCheck,
	PackageJSONCheck,
	AddPackageJSONScripts,
	installTools,
	DeprecrationsCheck,
	fixDocker,
	encodeApp,
	Plugins{}.RemoveOld,
	Plugins{}.CleanCache,
	Plugins{}.Reinstall,
	Plush,
}

func encodeApp(r *Runner) error {
	p := filepath.Join("config", "buffalo-app.toml")
	if _, err := os.Stat(p); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(r.App); err != nil {
		return err
	}
	return nil
}

func ask(q string) bool {
	if yesToAll {
		return true
	}

	fmt.Printf("? %s [y/n]\n", q)

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	text = strings.ToLower(strings.TrimSpace(text))
	return text == "y" || text == "yes"
}
