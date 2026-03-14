//go:build !embed_frontend

package front

import "io/fs"

func FS() fs.FS {
	return nil
}
