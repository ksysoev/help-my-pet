package media

import "sync"

// Collector provides functionality for managing and manipulating groups of media items.
// It is designed to handle concurrent access safely using an internal mutex. Media groups
// are stored and identified by unique group IDs, and each group can contain text and a
// collection of associated photo IDs.
type Collector struct {
	mediaGroups map[string]*Group
	mu          sync.Mutex
}

// Group represents a media group containing a caption text and an array of associated photo IDs.
// It stores the text of the group and a list of photo identifiers belonging to the group.
type Group struct {
	Text     string
	PhotoIDs []string
}

// NewCollector initializes and returns a pointer to a new Collector instance.
// It sets up an internal map to manage media groups and ensures thread-safe operations using a mutex.
// Returns *Collector, fully initialized and ready to manage media groups.
func NewCollector() *Collector {
	return &Collector{
		mediaGroups: make(map[string]*Group),
	}
}

// AddMediaGroup adds a text and photo ID to the specified media group or creates a new group if it does not exist.
// It locks the internal state for thread safety. The text is updated if non-empty, and the photo ID is appended.
// groupID specifies the unique identifier for the media group. text is the optional caption for the media.
// photoID is the identifier of the photo to add. Returns no value and does not produce errors.
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

// FinishMediaGroup retrieves and removes the media group associated with the given groupID from the collector.
// It deletes the group from internal storage to release resources and ensure subsequent operations receive fresh data.
// groupID specifies the unique identifier for the media group to finalize.
// Returns the removed Group object, or an empty Group object if the groupID does not exist.
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
