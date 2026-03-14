//go:build embed_frontend

package front

import (
	"embed"
	"io/fs"
)

//go:embed ../../web/dist/**
var distFS embed.FS

func GetFS() fs.FS {
	return distFS
}
