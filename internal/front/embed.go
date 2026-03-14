//go:build embed_frontend

package front

import (
	"io/fs"

	"github.com/gmlazutin/comparch-lab-3mod-3/web"
)

func FS() fs.FS {
	FS, err := fs.Sub(web.FS, "dist")
	if err != nil {
		panic("front: embedding enabled, however, \"dist\" path can not be reached: " + err.Error())
	}
	return FS
}
