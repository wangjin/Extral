//go:build windows

package ffmpeg

import "embed"

//go:embed all:bin
var EmbeddedBinaries embed.FS
