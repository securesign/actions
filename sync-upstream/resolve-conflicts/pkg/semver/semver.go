package semver

import (
	"strings"

	modsemver "golang.org/x/mod/semver"
)

func Compare(a, b string) int {
	return modsemver.Compare(Canonical(a), Canonical(b))
}

func Canonical(v string) string {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}

func Newer(a, b string) string {
	if Compare(a, b) >= 0 {
		return a
	}
	return b
}

func ClampToMax(version, ceiling string) string {
	if ceiling == "" {
		return version
	}
	if Compare(version, ceiling) > 0 {
		return ceiling
	}
	return version
}
