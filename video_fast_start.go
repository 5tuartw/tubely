package main

import (
	"os/exec"
	"bytes"
	"fmt"
)

func processvideoForFastStart(filePath string) (string, error) {
	newFilePath := filePath+".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", newFilePath)
	//var out bytes.Buffer
	//cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %w, stderr: %s", err, stderr.String())
	}

	return newFilePath, nil
}