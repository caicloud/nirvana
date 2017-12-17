package people

import "github.com/caicloud/nirvana/examples/swapi/pkg/model"

type ById []model.Person

func (r ById) Len() int {
	return len(r)
}

func (r ById) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r ById) Less(i, j int) bool {
	return r[i].Id < r[j].Id
}
