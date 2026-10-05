package tools

import "dvah/schemas"
import "encoding/json"
import _ "embed"

//go:embed Permissions.json
var permissions_json []byte

var PermissionsSchema []schemas.Tool

func init() {

	schema := make([]schemas.Tool, 0)
	err    := json.Unmarshal(permissions_json, &schema)

	if err == nil {
		PermissionsSchema = schema
	} else {
		panic(err)
	}

}
