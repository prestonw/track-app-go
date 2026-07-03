package ui

// buildVersion is set at compile time via -ldflags. Falls back when unset.
var buildVersion = "dev"

func BuildVersion() string {
	if buildVersion == "" {
		return "dev"
	}
	return buildVersion
}