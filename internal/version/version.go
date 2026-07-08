package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Version 是引擎的完整语义版本号。
// 打包时通过 -ldflags -X 注入，开发期使用包内默认值。
var Version = "v0.4.5"

// MinCompatibleVersion 是当前版本能够兼容对接的最低引擎版本（含）。
var MinCompatibleVersion = "v0.4.5"

// SemVer 表示解析后的语义版本号。
type SemVer struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion 将 "v0.3.0" 格式的版本字符串解析为 SemVer。
func ParseVersion(v string) (SemVer, error) {
	s := strings.TrimPrefix(v, "v")
	s = strings.TrimPrefix(s, "V")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return SemVer{}, fmt.Errorf("invalid version format: %s (expected vMAJOR.MINOR.PATCH)", v)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid major version in %s: %w", v, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid minor version in %s: %w", v, err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return SemVer{}, fmt.Errorf("invalid patch version in %s: %w", v, err)
	}
	return SemVer{Major: major, Minor: minor, Patch: patch}, nil
}

// Check 检查 engineVersion 是否与 currentVersion 兼容。
func Check(currentVersion, engineVersion string) (compatible bool, message string) {
	cv, err := ParseVersion(currentVersion)
	if err != nil {
		return false, fmt.Sprintf("cannot parse current version %s: %v", currentVersion, err)
	}
	ev, err := ParseVersion(engineVersion)
	if err != nil {
		return false, fmt.Sprintf("cannot parse engine version %s: %v", engineVersion, err)
	}

	if cv.Major != ev.Major {
		return false, fmt.Sprintf("version mismatch: Engine v%d.%d.%d is incompatible with %s (major version %d != %d)", ev.Major, ev.Minor, ev.Patch, currentVersion, ev.Major, cv.Major)
	}
	if cv.Minor != ev.Minor {
		return false, fmt.Sprintf("version mismatch: Engine v%d.%d.%d is incompatible with %s (minor version %d != %d)", ev.Major, ev.Minor, ev.Patch, currentVersion, ev.Minor, cv.Minor)
	}
	if ev.Patch < cv.Patch {
		return false, fmt.Sprintf("version mismatch: Engine v%d.%d.%d is older than required %s (patch %d < %d)", ev.Major, ev.Minor, ev.Patch, currentVersion, ev.Patch, cv.Patch)
	}

	return true, fmt.Sprintf("compatible (current=%s, engine=%s)", currentVersion, engineVersion)
}



