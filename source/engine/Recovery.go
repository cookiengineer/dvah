package engine

import "encoding/json"
import "fmt"
import "os"
import "path/filepath"
import "strings"

type Recovery struct {
	Sandbox string `json:"sandbox"`
}

func NewRecovery(sandbox string) *Recovery {

	return &Recovery{
		Sandbox: sandbox,
	}

}

func (recovery *Recovery) BackupSession(session *Session) error {

	bytes, err1 := json.MarshalIndent(session, "", "\t")

	if err1 == nil {

		path := filepath.Join(recovery.Sandbox, ".dvah", "session.json")
		err2 := os.MkdirAll(filepath.Dir(path), 0755)

		if err2 == nil {

			err3 := os.WriteFile(path, bytes, 0666)

			if err3 == nil {
				return nil
			} else {
				return err3
			}

		} else {
			return err2
		}

	} else {
		return err1
	}

}

func (recovery *Recovery) HasBackup() bool {

	path      := filepath.Join(recovery.Sandbox, ".dvah", "session.json")
	stat, err1 := os.Stat(path)

	if err1 == nil && stat.IsDir() == false {
		return true
	}

	return false

}

func (recovery *Recovery) RestoreSession() *Session {

	path        := filepath.Join(recovery.Sandbox, ".dvah", "session.json")
	bytes, err1 := os.ReadFile(path)

	if err1 == nil {

		tmp  := Session{}
		err2 := json.Unmarshal(bytes, &tmp)

		if err2 == nil {
			return RestoreSession(recovery.Sandbox, tmp)
		} else {
			return nil
		}

	}

	return nil

}

func (recovery *Recovery) Snapshot(name string, raw any) error {

	if strings.Contains(name, ".") == false {

		bytes, err1 := json.MarshalIndent(raw, "", "\t")

		if err1 == nil {

			path := filepath.Join(recovery.Sandbox, ".dvah", "debug", fmt.Sprintf("%s.json", name))
			err2 := os.MkdirAll(filepath.Dir(path), 0755)

			if err2 == nil {

				err3 := os.WriteFile(path, bytes, 0666)

				if err3 == nil {
					return nil
				} else {
					return err3
				}

			} else {
				return err2
			}

		} else {
			return err1
		}

	} else {
		return fmt.Errorf("invalid name \"%s\"", name)
	}

}

func (recovery *Recovery) SnapshotBytes(name string, raw []byte) error {

	if strings.Contains(name, ".") == false {

		var tmp interface{}

		err0 := json.Unmarshal(raw, &tmp)

		if err0 == nil {

			bytes, err1 := json.MarshalIndent(tmp, "", "\t")

			if err1 == nil {

				path := filepath.Join(recovery.Sandbox, ".dvah", "debug", fmt.Sprintf("%s.json", name))
				err2 := os.MkdirAll(filepath.Dir(path), 0755)

				if err2 == nil {

					err3 := os.WriteFile(path, bytes, 0666)

					if err3 == nil {
						return nil
					} else {
						return err3
					}

				} else {
					return err2
				}

			} else {
				return err1
			}

		} else {
			return err0
		}

	} else {
		return fmt.Errorf("invalid name \"%s\"", name)
	}

}
