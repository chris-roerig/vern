package version

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/chris/vern/internal/config"
)

type VersionInfo struct {
	Full  string
	Major int
	Minor int
	Patch int
}

func ParseVersion(s string) (VersionInfo, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return VersionInfo{}, fmt.Errorf("invalid version format: %s", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return VersionInfo{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return VersionInfo{}, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return VersionInfo{}, err
	}
	return VersionInfo{
		Full:  s,
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

func (v VersionInfo) Compare(other VersionInfo) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

func FetchAvailableVersions(lang *config.Language) ([]VersionInfo, error) {
	resp, err := http.Get(lang.VersionSource.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d fetching versions", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(lang.VersionSource.VersionRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid version regex: %w", err)
	}

	matches := re.FindAllStringSubmatch(string(body), -1)
	seen := make(map[string]bool)
	var versions []VersionInfo

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		verStr := match[1]
		if seen[verStr] {
			continue
		}
		seen[verStr] = true
		vi, err := ParseVersion(verStr)
		if err != nil {
			continue
		}
		versions = append(versions, vi)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Compare(versions[j]) < 0
	})

	return versions, nil
}

func ResolveVersion(lang *config.Language, versionArg string) (string, error) {
	available, err := FetchAvailableVersions(lang)
	if err != nil {
		return "", err
	}
	if len(available) == 0 {
		return "", fmt.Errorf("no versions found for %s", lang.Name)
	}

	if versionArg == "" {
		return available[len(available)-1].Full, nil
	}

	if _, err := ParseVersion(versionArg); err == nil {
		for _, v := range available {
			if v.Full == versionArg {
				return versionArg, nil
			}
		}
		return "", fmt.Errorf("version %s not found", versionArg)
	}

	parts := strings.Split(versionArg, ".")
	if len(parts) == 1 {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", fmt.Errorf("invalid version: %s", versionArg)
		}
		var candidates []VersionInfo
		for _, v := range available {
			if v.Major == major {
				candidates = append(candidates, v)
			}
		}
		if len(candidates) == 0 {
			return "", fmt.Errorf("no versions for major %s", versionArg)
		}
		return candidates[len(candidates)-1].Full, nil
	} else if len(parts) == 2 {
		major, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", fmt.Errorf("invalid version: %s", versionArg)
		}
		minor, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", fmt.Errorf("invalid version: %s", versionArg)
		}
		var candidates []VersionInfo
		for _, v := range available {
			if v.Major == major && v.Minor == minor {
				candidates = append(candidates, v)
			}
		}
		if len(candidates) == 0 {
			return "", fmt.Errorf("no versions for %s", versionArg)
		}
		return candidates[len(candidates)-1].Full, nil
	}

	return "", fmt.Errorf("invalid version: %s", versionArg)
}
