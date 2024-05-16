package configmanager

import (
	_ "embed"
	"strings"
)

//go:generate cp -r ../../latest ./resources/version
//go:embed resources/version
var version string

func getVersion() string {
	return strings.TrimSpace(version)
}
