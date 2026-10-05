package fs

import "os"
import "path/filepath"

func HasDvahSession(folder string) bool {

	tmp         := filepath.Join(folder, ".dvah", "session.json")
	stat, err := os.Stat(tmp)

	if err == nil && stat.IsDir() == false {
		return true
	}

	return false

}

