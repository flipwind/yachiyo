package endpoint

import "maps"

// Capability defines a capability map that a *Client* has.
type Capability struct {
	capMap map[string]bool
}

const CapMessageText = "message.text"

func AvailableCapability() []string {
	return []string{CapMessageText}
}

func initialize() map[string]bool {
	// available capabilities
	li := AvailableCapability()

	finalMap := make(map[string]bool)
	for _, l := range li {
		finalMap[l] = false
	}

	return finalMap
}

func NewCapability() *Capability {
	return &Capability{
		capMap: initialize(),
	}
}

func (c *Capability) Reset() {
	c.capMap = initialize()
}

// If capability exist, then set it to true.
// If not, then return false and do nothing.
func (c *Capability) Set(cap string) (exist bool) {
	_, ok := c.capMap[cap]
	if !ok {
		return false
	}

	c.capMap[cap] = true
	return true
}

func (c *Capability) Has(cap string) bool {
	return c.capMap[cap]
}

func (c *Capability) EnableCounts() int {
	count := 0
	for _, cap := range c.capMap {
		if cap {
			count++
		}
	}

	return count
}

func (c *Capability) Clone() Capability {
	return Capability{
		capMap: maps.Clone(c.capMap),
	}
}
