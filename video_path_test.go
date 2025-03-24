package main

import (
	"fmt"
	"testing"
)

func TestGetVideoAspectRatio(t *testing.T) {
	// Replace with the path to a test video file on your system
	testVideoPath := "./samples/boots-video-horizontal.mp4"
	
	fmt.Println("Testing getVideoAspectRatio with file:", testVideoPath)
	
	ratio, err := getVideoAspectRatio(testVideoPath)
	if err != nil {
		t.Fatalf("Error getting aspect ratio: %v", err)
	}
	
	fmt.Println("Aspect ratio:", ratio)
	t.Logf("Aspect ratio: %s", ratio)
}