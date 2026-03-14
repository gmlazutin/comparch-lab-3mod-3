//go:build !embed_frontend

package front

import "io/fs"

func GetFS() fs.FS {
	return nil
}
