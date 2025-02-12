package media

import "sync"

type Collector struct {
	mediaGroups map[string]*Group
	mu          sync.Mutex
}

type Group struct {
	Text     string
	PhotoIDs []string
}

func NewCollector() *Collector {
	return &Collector{
		mediaGroups: make(map[string]*Group),
	}
}

func (c *Collector) AddMediaGroup(groupID string, text string, photoID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	group, ok := c.mediaGroups[groupID]
	if !ok {
		group = &Group{}
	}

	if text != "" {
		group.Text = text
	}

	group.PhotoIDs = append(group.PhotoIDs, photoID)

	c.mediaGroups[groupID] = group
}

func (c *Collector) FinishMediaGroup(groupID string) *Group {
	c.mu.Lock()
	defer c.mu.Unlock()

	group, ok := c.mediaGroups[groupID]
	if !ok {
		return &Group{}
	}

	delete(c.mediaGroups, groupID)

	return group
}
