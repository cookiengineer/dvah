package adapters

import "dvah/types"

var Registry []types.Adapter

func init() {

	Registry = []types.Adapter{
		&DeepSeek{},
		&NoTransform{},
	}

}
