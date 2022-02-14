package bins

import (
	"embed"
)

//go:embed stub-*_*
var FS embed.FS

const (
	AMD64BinaryName = "stub-binary-darwin_amd64"
	ARM64BinaryName = "stub-binary-darwin_arm64"
)
