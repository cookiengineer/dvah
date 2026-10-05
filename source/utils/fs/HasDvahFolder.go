package fs

import "os"
import "path/filepath"

func HasDvahFolder(folder string) bool {

	tmp         := filepath.Join(folder, ".dvah")
	stat, err := os.Stat(tmp)

	if err == nil && stat.IsDir() {
		return true
	}

	return false

}

