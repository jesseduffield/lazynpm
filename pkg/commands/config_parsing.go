package commands

import (
	"io"
	"io/ioutil"

	"github.com/buger/jsonparser"
)

func unescape(b []byte) string {
	buf := make([]byte, 0, len(b))
	buf, err := jsonparser.Unescape(b, buf)
	if err != nil {
		return string(b)
	}
	return string(buf)
}

func UnmarshalPackageConfig(r io.Reader) (*PackageConfig, error) {
	configData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	pkgConfig := &PackageConfig{}

	type stringMapping struct {
		path []string
		ptr  *string
	}

	parseStrings := func(data []byte, mappings []stringMapping) {
		for _, mapping := range mappings {
			value, err := jsonparser.GetString(data, mapping.path...)
			if err == nil {
				*mapping.ptr = value
			}
		}
	}

	parseStrings(configData,
		[]stringMapping{
			{
				path: []string{"name"},
				ptr:  &pkgConfig.Name,
			},
			{
				path: []string{"version"},
				ptr:  &pkgConfig.Version,
			},
			{
				path: []string{"license"},
				ptr:  &pkgConfig.License,
			},
			{
				path: []string{"description"},
				ptr:  &pkgConfig.Description,
			},
			{
				path: []string{"homepage"},
				ptr:  &pkgConfig.Homepage,
			},
			{
				path: []string{"main"},
				ptr:  &pkgConfig.Main,
			},
			{
				path: []string{"engines", "node"},
				ptr:  &pkgConfig.Engines.Node,
			},
			{
				path: []string{"engines", "npm"},
				ptr:  &pkgConfig.Engines.Npm,
			},
			{
				path: []string{"repository"},
				ptr:  &pkgConfig.Repository.SingleLine,
			},
			{
				path: []string{"repository", "type"},
				ptr:  &pkgConfig.Repository.Type,
			},
			{
				path: []string{"repository", "url"},
				ptr:  &pkgConfig.Repository.Url,
			},
			{
				path: []string{"author"},
				ptr:  &pkgConfig.Author.SingleLine,
			},
			{
				path: []string{"author", "email"},
				ptr:  &pkgConfig.Author.Email,
			},
			{
				path: []string{"author", "name"},
				ptr:  &pkgConfig.Author.Name,
			},
			{
				path: []string{"author", "url"},
				ptr:  &pkgConfig.Author.Url,
			},
			{
				path: []string{"bugs"},
				ptr:  &pkgConfig.Bugs.Url,
			},
			{
				path: []string{"bugs", "url"},
				ptr:  &pkgConfig.Bugs.Url,
			},
		},
	)

	for _, mapping := range []struct {
		field string
		ptr   *bool
	}{
		{
			field: "deprecated",
			ptr:   &pkgConfig.Deprecated,
		},
		{
			field: "private",
			ptr:   &pkgConfig.Private,
		},
	} {
		value, err := jsonparser.GetBoolean(configData, mapping.field)
		if err == nil {
			*mapping.ptr = value
		}
	}

	for _, mapping := range []struct {
		field string
		ptr   *[]string
	}{
		{
			field: "files",
			ptr:   &pkgConfig.Files,
		},
		{
			field: "keywords",
			ptr:   &pkgConfig.Keywords,
		},
		{
			field: "os",
			ptr:   &pkgConfig.Os,
		},
		{
			field: "cpu",
			ptr:   &pkgConfig.Cpu,
		},
		{
			field: "bundleDependencies",
			ptr:   &pkgConfig.BundledDependencies,
		},
		{
			// both this spelling and the spelling above are honoured
			field: "bundledDependencies",
			ptr:   &pkgConfig.BundledDependencies,
		},
	} {
		value, dataType, _, err := jsonparser.Get(configData, mapping.field)
		if err != nil {
			if err == jsonparser.KeyPathNotFoundError {
				continue
			}
			return nil, err
		}
		switch dataType {
		case jsonparser.Array:
			_, _ = jsonparser.ArrayEach(value, func(innerValue []byte, dataType jsonparser.ValueType, offset int, err error) {
				if dataType == jsonparser.String {
					*mapping.ptr = append(*mapping.ptr, unescape(innerValue))
				}
			})
		case jsonparser.String:
			*mapping.ptr = append(*mapping.ptr, string(value))
		}
	}

	for _, mapping := range []struct {
		field string
		ptr   *map[string]string
	}{
		{
			field: "scripts",
			ptr:   &pkgConfig.Scripts,
		},
		{
			field: "directories",
			ptr:   &pkgConfig.Directories,
		},
		{
			field: "dependencies",
			ptr:   &pkgConfig.Dependencies,
		},
		{
			field: "devDependencies",
			ptr:   &pkgConfig.DevDependencies,
		},
		{
			field: "peerDependencies",
			ptr:   &pkgConfig.PeerDependencies,
		},
		{
			field: "optionalDependencies",
			ptr:   &pkgConfig.OptionalDependencies,
		},
	} {
		value, dataType, _, err := jsonparser.Get(configData, mapping.field)
		if err != nil {
			if err == jsonparser.KeyPathNotFoundError {
				continue
			}
			return nil, err
		}
		switch dataType {
		case jsonparser.Object:
			_ = jsonparser.ObjectEach(value, func(key []byte, innerValue []byte, dataType jsonparser.ValueType, offset int) error {
				if dataType == jsonparser.String {
					if *mapping.ptr == nil {
						*mapping.ptr = map[string]string{}
					}
					(*mapping.ptr)[string(key)] = unescape(innerValue)
				}
				return nil
			})
		}
	}
	value, dataType, _, err := jsonparser.Get(configData, "contributors")
	if err != nil {
		if err != jsonparser.KeyPathNotFoundError {
			return nil, err
		}
	} else {
		switch dataType {
		case jsonparser.Array:
			_, _ = jsonparser.ArrayEach(value, func(innerValue []byte, dataType jsonparser.ValueType, offset int, err error) {
				switch dataType {
				case jsonparser.String:
					pkgConfig.Contributors = append(pkgConfig.Contributors, Author{SingleLine: unescape(innerValue)})
				case jsonparser.Object:
					contributor := Author{}
					parseStrings(innerValue,
						[]stringMapping{
							{
								path: []string{""},
								ptr:  &contributor.SingleLine,
							},
							{
								path: []string{"email"},
								ptr:  &contributor.Email,
							},
							{
								path: []string{"name"},
								ptr:  &contributor.Name,
							},
							{
								path: []string{"url"},
								ptr:  &contributor.Url,
							},
						},
					)
					pkgConfig.Contributors = append(pkgConfig.Contributors, contributor)
				}
			})
		}
	}

	return pkgConfig, nil
}
