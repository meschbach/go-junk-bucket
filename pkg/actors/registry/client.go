package registry

import "github.com/meschbach/go-junk-bucket/pkg/actors"

type ControllingClient struct {
	bif     actors.Runtime
	control actors.Pid
}

func NewControllingClient(bif actors.Runtime, registry actors.Pid) ControllingClient {
	return ControllingClient{bif: bif, control: registry}
}

func (c *ControllingClient) Register(name string, who actors.Pid) {
	c.bif.Tell(c.control, Register{
		Name: name,
		Who:  who,
	})
}

func (c *ControllingClient) Unregister(name string, who actors.Pid) {
	c.bif.Tell(c.control, Unregister{
		Name: name,
		Who:  who,
	})
}

func (c *ControllingClient) Lookup(name string, tell actors.Pid) {
	c.bif.Tell(c.control, Lookup{
		Tell: tell,
		Name: name,
	})
}

func (c *ControllingClient) Ref(name string) NamedRef {
	//TODO: Might want to use proxy actor to ensure no schenangians
	return NamedRef{
		name:  name,
		query: c.control,
	}
}

type QueryClient struct {
	bif   actors.Runtime
	query actors.Pid
}

func NewQueryClient(bif actors.Runtime, query actors.Pid) QueryClient {
	return QueryClient{
		bif:   bif,
		query: query,
	}
}

func (c *QueryClient) Lookup(name string, tell actors.Pid) {
	c.bif.Tell(c.query, Lookup{
		Tell: tell,
		Name: name,
	})
}

func (c *QueryClient) Find(name string) actors.Pid {
	result := actors.CallService[Queryable, LookupResult](c.bif, c.query, &LookupRPC{
		Name: name,
	})
	if !result.Found {
		c.bif.Log().Fatal("failed to find %s against %#v", name, c.query)
	}
	return result.Who
}

type Queryable interface {
	Lookup(runtime actors.Runtime, name string) LookupResult
}

type LookupRPC struct {
	Name string
}

func (l *LookupRPC) Invoke(bif actors.Runtime, state Queryable) LookupResult {
	return state.Lookup(bif, l.Name)
}
