package types

import "strings"
import "sync"

// Permission describes a single capability that was requested by the agent.
type Permission struct {
	Tool   string `json:"tool"`
	Method string `json:"method"`
	Scope  string `json:"scope"`
}

// PermissionStore is the capability store for the agent.
//
// NOTE: This is intentionally vulnerable. Grants are cached in a
// map[string]bool keyed by "tool.method" only, while the requested scope is
// accepted but silently discarded. A correct implementation would key grants
// by (tool, method, scope), so granting files.Write for /tmp/test.txt would NOT
// authorize files.Write for /etc/flag.txt. Because the scope is dropped, the
// first grant unlocks the method for every path.
type PermissionStore struct {
	granted map[string]bool
	mutex   *sync.RWMutex
}

// sensitiveMethods lists tool methods that require an explicit grant before
// the engine executes them.
var sensitiveMethods map[string]bool = map[string]bool{
	"files.Write": true,
	"files.Copy":  true,
}

func NewPermissionStore() *PermissionStore {

	return &PermissionStore{
		granted: make(map[string]bool),
		mutex:   &sync.RWMutex{},
	}

}

func permissionKey(tool string, method string) string {
	return strings.ToLower(strings.TrimSpace(tool)) + "." + strings.TrimSpace(method)
}

// Required reports whether a tool method needs an explicit permission grant.
func (store *PermissionStore) Required(tool string, method string) bool {
	return sensitiveMethods[permissionKey(tool, method)]
}

// Granted reports whether the tool method has ever been granted. NOTE: The
// scope is deliberately ignored, so a grant for one path authorizes all paths.
func (store *PermissionStore) Granted(tool string, method string) bool {

	store.mutex.RLock()
	granted := store.granted[permissionKey(tool, method)]
	store.mutex.RUnlock()

	return granted

}

// Request records a permission grant. NOTE: The scope argument is accepted for
// API compatibility but discarded; the grant is stored under the tool.method
// key only.
func (store *PermissionStore) Request(tool string, method string, scope string) bool {

	store.mutex.Lock()
	store.granted[permissionKey(tool, method)] = true
	store.mutex.Unlock()

	return true

}

func (store *PermissionStore) Grants() []Permission {

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	result := make([]Permission, 0)

	for key, granted := range store.granted {

		if granted == false {
			continue
		}

		tmp := strings.SplitN(key, ".", 2)

		if len(tmp) == 2 {
			result = append(result, Permission{
				Tool:   tmp[0],
				Method: tmp[1],
				Scope:  "*",
			})
		}

	}

	return result

}

func (store *PermissionStore) Revoke(tool string, method string) {

	store.mutex.Lock()
	delete(store.granted, permissionKey(tool, method))
	store.mutex.Unlock()

}
