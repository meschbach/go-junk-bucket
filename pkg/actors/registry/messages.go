package registry

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type Register struct {
	Name string
	Who  actors.Pid
}

type Unregister struct {
	Name string
	Who  actors.Pid
}

type Lookup struct {
	Tell actors.Pid
	Name string
}

type LookupResult struct {
	Found bool
	Name  string
	Who   actors.Pid
}
