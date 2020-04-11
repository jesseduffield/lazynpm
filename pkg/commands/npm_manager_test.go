package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalPackageConfig(t *testing.T) {
	type scenario struct {
		filename              string
		expectedPackageConfig *PackageConfig
	}

	scenarios := []scenario{
		{
			"1.json",
			&PackageConfig{Name: "body", Version: "5.1.0", License: "", Private: false, Description: "Body parsing", Files: []string(nil), Keywords: []string{}, Os: []string(nil), Cpu: []string(nil), Main: "index", Engines: struct {
				Node string
				Npm  string
			}{Node: "", Npm: ""}, Scripts: map[string]string{"test": "node ./test/index.js"}, Repository: Repository{Type: "", Url: "", SingleLine: "git://github.com/Raynos/body.git"}, Author: Author{Name: "", Email: "", Url: "", SingleLine: "Raynos <raynos2@gmail.com>"}, Contributors: []Author{Author{Name: "Jake Verbaten", Email: "", Url: "", SingleLine: ""}}, Bugs: struct {
				Url string "json:\"url\""
			}{Url: "https://github.com/Raynos/body/issues"}, Deprecated: false, Homepage: "https://github.com/Raynos/body", Directories: map[string]string(nil), Dependencies: map[string]string{"continuable-cache": "^0.3.1", "error": "^7.0.0", "raw-body": "~1.1.0", "safe-json-parse": "~1.0.1"}, DevDependencies: map[string]string{"after": "~0.7.0", "hammock": "^1.0.0", "process": "~0.5.1", "send-data": "~1.0.1", "tape": "~2.3.0", "test-server": "~0.1.3"}, PeerDependencies: map[string]string(nil), OptionalDependencies: map[string]string(nil), BundledDependencies: []string(nil)},
		},
		{
			"2.json",
			&PackageConfig{Name: "lodash", Version: "4.17.5", License: "MIT", Private: false, Description: "Lodash modular utilities.", Files: []string(nil), Keywords: []string{"modules, stdlib, util"}, Os: []string(nil), Cpu: []string(nil), Main: "lodash.js", Engines: struct {
				Node string
				Npm  string
			}{Node: "", Npm: ""}, Scripts: map[string]string{"test": "echo \"See https://travis-ci.org/lodash-archive/lodash-cli for testing details.\""}, Repository: Repository{Type: "", Url: "", SingleLine: "lodash/lodash"}, Author: Author{Name: "", Email: "", Url: "", SingleLine: "John-David Dalton <john.david.dalton@gmail.com> (http://allyoucanleet.com/)"}, Contributors: []Author{Author{Name: "", Email: "", Url: "", SingleLine: "John-David Dalton <john.david.dalton@gmail.com> (http://allyoucanleet.com/)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Mathias Bynens <mathias@qiwi.be> (https://mathiasbynens.be/)"}}, Bugs: struct {
				Url string "json:\"url\""
			}{Url: ""}, Deprecated: false, Homepage: "https://lodash.com/", Directories: map[string]string(nil), Dependencies: map[string]string(nil), DevDependencies: map[string]string(nil), PeerDependencies: map[string]string(nil), OptionalDependencies: map[string]string(nil), BundledDependencies: []string(nil)},
		},
		{
			"3.json",
			&PackageConfig{Name: "moment-range", Version: "2.2.0", License: "", Private: false, Description: "Fancy date ranges for Moment.js", Files: []string(nil), Keywords: []string(nil), Os: []string(nil), Cpu: []string(nil), Main: "./dist/moment-range", Engines: struct {
				Node string
				Npm  string
			}{Node: "*", Npm: ""}, Scripts: map[string]string{"build": "grunt es6transpiler replace umd uglify", "jsdoc": "jsdoc -c .jsdoc", "test": "grunt mochaTest"}, Repository: Repository{Type: "git", Url: "https://git@github.com/gf3/moment-range.git", SingleLine: ""}, Author: Author{Name: "", Email: "", Url: "", SingleLine: "Gianni Chiappetta <gianni@runlevel6.org> (http://butt.zone)"}, Contributors: []Author{Author{Name: "", Email: "", Url: "", SingleLine: "Adam Biggs <adam.biggs@lightmaker.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Matt Patterson <matt@reprocessed.org> (http://reprocessed.org/)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Stuart Kelly <stuart.leigh83@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Kevin Ross <kevin.ross@alienfast.com> (http://www.alienfast.com)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Scott Hovestadt <scott.hovestadt@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Nebel <nebel08@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Aristide Niyungeko <niyungeko@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Tymon Tobolski <i@teamon.eu> (http://teamon.eu)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Bradley Ayers <bradley.ayers@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Thomas Walpole <twalpole@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Daniel Sarfati <daniel@knockrentals.com>"}}, Bugs: struct {
				Url string "json:\"url\""
			}{Url: "https://github.com/gf3/moment-range/issues"}, Deprecated: false, Homepage: "https://github.com/gf3/moment-range", Directories: map[string]string{"lib": "./lib"}, Dependencies: map[string]string(nil), DevDependencies: map[string]string{"grunt": "~0.4.1", "grunt-cli": "^0.1.13", "grunt-contrib-uglify": "^0.6.0", "grunt-es6-transpiler": "^1.0.2", "grunt-mocha-test": "~0.7.0", "grunt-text-replace": "^0.4.0", "grunt-umd": "^2.3.3", "jsdoc": "^3.3.0", "mocha": "^2.1.0", "moment": ">= 1", "should": "^5.0.1"}, PeerDependencies: map[string]string{"moment": ">= 1"}, OptionalDependencies: map[string]string(nil), BundledDependencies: []string(nil)},
		},
		{
			"4.json",
			&PackageConfig{Name: "moment-range", Version: "2.2.0", License: "", Private: false, Description: "Fancy date ranges for Moment.js", Files: []string(nil), Keywords: []string(nil), Os: []string(nil), Cpu: []string(nil), Main: "./dist/moment-range", Engines: struct {
				Node string
				Npm  string
			}{Node: "*", Npm: ""}, Scripts: map[string]string{"build": "grunt es6transpiler replace umd uglify", "jsdoc": "jsdoc -c .jsdoc", "test": "grunt mochaTest"}, Repository: Repository{Type: "git", Url: "https://git@github.com/gf3/moment-range.git", SingleLine: ""}, Author: Author{Name: "", Email: "", Url: "", SingleLine: "Gianni Chiappetta <gianni@runlevel6.org> (http://butt.zone)"}, Contributors: []Author{Author{Name: "", Email: "", Url: "", SingleLine: "Adam Biggs <adam.biggs@lightmaker.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Matt Patterson <matt@reprocessed.org> (http://reprocessed.org/)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Stuart Kelly <stuart.leigh83@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Kevin Ross <kevin.ross@alienfast.com> (http://www.alienfast.com)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Scott Hovestadt <scott.hovestadt@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Nebel <nebel08@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Aristide Niyungeko <niyungeko@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Tymon Tobolski <i@teamon.eu> (http://teamon.eu)"}, Author{Name: "", Email: "", Url: "", SingleLine: "Bradley Ayers <bradley.ayers@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Thomas Walpole <twalpole@gmail.com>"}, Author{Name: "", Email: "", Url: "", SingleLine: "Daniel Sarfati <daniel@knockrentals.com>"}}, Bugs: struct {
				Url string "json:\"url\""
			}{Url: "https://github.com/gf3/moment-range/issues"}, Deprecated: false, Homepage: "https://github.com/gf3/moment-range", Directories: map[string]string{"lib": "./lib"}, Dependencies: map[string]string(nil), DevDependencies: map[string]string{"grunt": "~0.4.1", "grunt-cli": "^0.1.13", "grunt-contrib-uglify": "^0.6.0", "grunt-es6-transpiler": "^1.0.2", "grunt-mocha-test": "~0.7.0", "grunt-text-replace": "^0.4.0", "grunt-umd": "^2.3.3", "jsdoc": "^3.3.0", "mocha": "^2.1.0", "moment": ">= 1", "should": "^5.0.1"}, PeerDependencies: map[string]string{"moment": ">= 1"}, OptionalDependencies: map[string]string(nil), BundledDependencies: []string(nil)},
		},
	}

	for _, s := range scenarios {
		file, err := os.OpenFile(filepath.Join("testfiles", s.filename), os.O_RDONLY, 0644)
		if err != nil {
			panic(err)
		}

		config, err := UnmarshalPackageConfig(file)

		assert.NoError(t, err)
		assert.EqualValues(t, s.expectedPackageConfig, config)
	}
}
