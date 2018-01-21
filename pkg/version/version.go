package version

import (
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/drausin/libri/libri/common/errors"
)

// Current contains the current build info.
var Current BuildInfo

// these variables are populated by ldflags during builds and fall back to population from git repo
// when they're not set (e.g., during tests)
var (
	// GitBranch is the current git branch
	GitBranch string

	// GitRevision is the current git commit hash.
	GitRevision string

	// BuildDate is the date of the build.
	BuildDate string
)

var semverString = "0.1.0"

const (
	develop         = "develop"
	master          = "master"
	snapshot        = "snapshot"
	buildDateFormat = "2006-01-02" // ISO 8601 date format
)

var branchPrefixes = []string{
	"feature/",
	"release/",
	"bugfix/",
}

// BuildInfo contains info about the current build.
type BuildInfo struct {
	Version     semver.Version
	GitBranch   string
	GitRevision string
	BuildDate   string
}

func init() {
	Current = GetBuildInfo(GitBranch, GitRevision, BuildDate)
}

// GetBuildInfo gets the BuildInfo from build flags or local git repo info.
func GetBuildInfo(gitBranch, gitRevision, buildDate string) BuildInfo {
	wd, err := os.Getwd()
	errors.MaybePanic(err)
	g := git{dir: wd}

	if gitBranch == "" {
		gitBranch = g.Branch()
	}
	if gitRevision == "" {
		gitRevision, err = g.Commit()
		errors.MaybePanic(err)
	}
	if buildDate == "" {
		buildDate = time.Now().UTC().Format(buildDateFormat)
	}
	version := semver.MustParse(semverString)
	if gitBranch == master {
		// no pre-release tags to add
	} else if gitBranch == develop {
		version.Pre = []semver.PRVersion{{VersionStr: snapshot}}
	} else {
		version.Pre = []semver.PRVersion{{VersionStr: stripPrefixes(gitBranch)}}
	}
	return BuildInfo{
		Version:     version,
		GitBranch:   gitBranch,
		GitRevision: gitRevision,
		BuildDate:   buildDate,
	}
}

func stripPrefixes(branch string) string {
	for _, prefix := range branchPrefixes {
		if strings.HasPrefix(branch, prefix) {
			return strings.TrimPrefix(branch, prefix)
		}
	}
	return branch
}
