package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLayout(t *testing.T) {
	layout := NewLayout("test", 100, 50)
	assert.Equal(t, "test", layout.name)
	assert.Equal(t, 100, layout.width)
	assert.Equal(t, 50, layout.height)
	assert.NotNil(t, layout.regions)
}

func TestLayout_AddRegion(t *testing.T) {
	layout := NewLayout("test", 100, 50)
	region := NewRegion(0, 0, 50, 25, nil)
	layout.AddRegion("main", region)

	retrieved := layout.GetRegion("main")
	require.NotNil(t, retrieved)
	assert.Equal(t, 50, retrieved.width)
	assert.Equal(t, 25, retrieved.height)
}

func TestLayout_RemoveRegion(t *testing.T) {
	layout := NewLayout("test", 100, 50)
	region := NewRegion(0, 0, 50, 25, nil)
	layout.AddRegion("main", region)

	layout.RemoveRegion("main")
	retrieved := layout.GetRegion("main")
	assert.Nil(t, retrieved)
}

func TestLayout_SetSize(t *testing.T) {
	layout := NewLayout("test", 100, 50)
	layout.SetSize(200, 100)
	assert.Equal(t, 200, layout.width)
	assert.Equal(t, 100, layout.height)
}

func TestCreateDefaultLayout(t *testing.T) {
	layout := CreateDefaultLayout(100, 50)
	assert.NotNil(t, layout)
	assert.Equal(t, "default", layout.name)

	// Check that all expected regions exist
	assert.NotNil(t, layout.GetRegion("status"))
	assert.NotNil(t, layout.GetRegion("output"))
	assert.NotNil(t, layout.GetRegion("input"))
}

func TestCreateCompactLayout(t *testing.T) {
	layout := CreateCompactLayout(100, 50)
	assert.NotNil(t, layout)
	assert.Equal(t, "compact", layout.name)

	// Verify regions
	assert.NotNil(t, layout.GetRegion("status"))
	assert.NotNil(t, layout.GetRegion("output"))
	assert.NotNil(t, layout.GetRegion("input"))
}

func TestCreateSplitScreenLayout(t *testing.T) {
	layout := CreateSplitScreenLayout(100, 50)
	assert.NotNil(t, layout)
	assert.Equal(t, "split-screen", layout.name)

	// Verify split-screen regions
	assert.NotNil(t, layout.GetRegion("status"))
	assert.NotNil(t, layout.GetRegion("conversation"))
	assert.NotNil(t, layout.GetRegion("viewer"))
	assert.NotNil(t, layout.GetRegion("input"))
}

func TestCreateFocusLayout(t *testing.T) {
	layout := CreateFocusLayout(100, 50)
	assert.NotNil(t, layout)
	assert.Equal(t, "focus", layout.name)

	// Verify output takes full screen
	output := layout.GetRegion("output")
	require.NotNil(t, output)
	assert.Equal(t, 100, output.width)
	assert.Equal(t, 50, output.height)
}

func TestNewRegion(t *testing.T) {
	region := NewRegion(10, 20, 50, 30, nil)
	assert.Equal(t, 10, region.x)
	assert.Equal(t, 20, region.y)
	assert.Equal(t, 50, region.width)
	assert.Equal(t, 30, region.height)
	assert.False(t, region.border)
	assert.Equal(t, 0, region.padding)
}

func TestRegion_SetBorder(t *testing.T) {
	region := NewRegion(0, 0, 50, 25, nil)
	assert.False(t, region.border)
	region.SetBorder(true)
	assert.True(t, region.border)
}

func TestRegion_SetPadding(t *testing.T) {
	region := NewRegion(0, 0, 50, 25, nil)
	assert.Equal(t, 0, region.padding)
	region.SetPadding(2)
	assert.Equal(t, 2, region.padding)
}

func TestRegion_SetSize(t *testing.T) {
	region := NewRegion(0, 0, 50, 25, nil)
	region.SetSize(100, 50)
	assert.Equal(t, 100, region.width)
	assert.Equal(t, 50, region.height)
}

func TestRegion_SetPosition(t *testing.T) {
	region := NewRegion(0, 0, 50, 25, nil)
	region.SetPosition(10, 20)
	assert.Equal(t, 10, region.x)
	assert.Equal(t, 20, region.y)
}

// Test the max helper function
func TestMax(t *testing.T) {
	assert.Equal(t, 5, max(3, 5))
	assert.Equal(t, 10, max(10, 7))
	assert.Equal(t, 0, max(0, 0))
	assert.Equal(t, -1, max(-5, -1))
}
