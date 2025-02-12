package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector_AddMediaGroup(t *testing.T) {
	tests := []struct {
		name         string
		initialState map[string]*Group
		groupID      string
		text         string
		photoID      string
		wantGroup    *Group
	}{
		{
			name:         "new group with text and photo",
			initialState: map[string]*Group{},
			groupID:      "group1",
			text:         "some text",
			photoID:      "photo1",
			wantGroup: &Group{
				Text:     "some text",
				PhotoIDs: []string{"photo1"},
			},
		},
		{
			name: "update existing group with new photo",
			initialState: map[string]*Group{
				"group1": {Text: "existing text", PhotoIDs: []string{"photo1"}},
			},
			groupID: "group1",
			text:    "",
			photoID: "photo2",
			wantGroup: &Group{
				Text:     "existing text",
				PhotoIDs: []string{"photo1", "photo2"},
			},
		},
		{
			name: "update existing group with new text and photo",
			initialState: map[string]*Group{
				"group1": {Text: "existing text", PhotoIDs: []string{"photo1"}},
			},
			groupID: "group1",
			text:    "new text",
			photoID: "photo2",
			wantGroup: &Group{
				Text:     "new text",
				PhotoIDs: []string{"photo1", "photo2"},
			},
		},
		{
			name:         "new group with only photo",
			initialState: map[string]*Group{},
			groupID:      "group2",
			text:         "",
			photoID:      "photo3",
			wantGroup: &Group{
				Text:     "",
				PhotoIDs: []string{"photo3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			collector := NewCollector()
			collector.mediaGroups = tt.initialState

			// Act
			collector.AddMediaGroup(tt.groupID, tt.text, tt.photoID)

			// Assert
			gotGroup := collector.mediaGroups[tt.groupID]
			assert.NotNil(t, gotGroup)
			assert.Equal(t, tt.wantGroup.Text, gotGroup.Text)
			assert.Equal(t, tt.wantGroup.PhotoIDs, gotGroup.PhotoIDs)
		})
	}
}

func TestCollector_FinishMediaGroup(t *testing.T) {
	tests := []struct {
		name         string
		initialState map[string]*Group
		groupID      string
		wantGroup    *Group
		shouldDelete bool
	}{
		{
			name: "finish existing group",
			initialState: map[string]*Group{
				"group1": {Text: "some text", PhotoIDs: []string{"photo1", "photo2"}},
			},
			groupID: "group1",
			wantGroup: &Group{
				Text:     "some text",
				PhotoIDs: []string{"photo1", "photo2"},
			},
		},
		{
			name:         "finish non-existing group",
			initialState: map[string]*Group{},
			groupID:      "nonexistent",
			wantGroup:    &Group{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			collector := NewCollector()
			collector.mediaGroups = tt.initialState

			// Act
			gotGroup := collector.FinishMediaGroup(tt.groupID)

			// Assert
			assert.NotNil(t, gotGroup)
			assert.Equal(t, tt.wantGroup.Text, gotGroup.Text)
			assert.Equal(t, tt.wantGroup.PhotoIDs, gotGroup.PhotoIDs)

			_, exists := collector.mediaGroups[tt.groupID]
			assert.False(t, exists)
		})
	}
}
