package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	/*"runtime/debug"
	"time"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"*/
)

type VideoInfo struct {
	Streams []struct {
		CodecType          string `json:"codec_type"`
		Width              int    `json:"width"`
		Height             int    `json:"height"`
		DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %w, stderr: %s", err, stderr.String())
	}
	// Debug output
	outStr := out.String()
	//log.Printf("Raw output length: %d", len(outStr))
	//if len(outStr) > 0 {
	//	log.Printf("First 100 chars: %q", outStr[:min(100, len(outStr))])
	//}

	var info VideoInfo
	if err = json.Unmarshal(out.Bytes(), &info); err != nil {
		// Log the error and the position for debugging
		log.Printf("JSON error: %v", err)

		// Check if the output actually looks like JSON
		trimmed := strings.TrimSpace(outStr)
		if !strings.HasPrefix(trimmed, "{") {
			return "", fmt.Errorf("invalid output format from ffprobe: %s", trimmed[:min(50, len(trimmed))])
		}

		return "", fmt.Errorf("JSON unmarshal error: %w", err)
	}

	for _, stream := range info.Streams {
		if stream.CodecType == "video" {
			if stream.Width > 0 && stream.Height > 0 {
				// Calculate the aspect ratio
				aspectRatio := float64(stream.Width) / float64(stream.Height)

				// Determine the aspect ratio category
				if aspectRatio > 1.7 && aspectRatio < 1.8 { // Close to 16:9
					return "16:9", nil
				} else if aspectRatio > 0.55 && aspectRatio < 0.57 { // Close to 9:16
					return "9:16", nil
				} else {
					return "other", nil
				}
			}
		}
	}

	// No video stream found
	return "", fmt.Errorf("no video stream found in file")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func processvideoForFastStart(filePath string) (string, error) {
	newFilePath := filePath + ".processing"
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

/*func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignedClient := s3.NewPresignClient(s3Client)
	request, err := presignedClient.PresignGetObject(context.Background(),
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	fmt.Println("dbVideoToSignedVideo called with video ID:", video.ID)
	debug.PrintStack()

	if video.VideoURL == nil {
		return video, nil
	}

	videoParams := strings.Split(*video.VideoURL, ",")
	if len(videoParams) != 2 {
		return video, fmt.Errorf("could not parse video information for presigned URL from VideoURL")
	}
	bucket := videoParams[0]
	key := videoParams[1]

	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, (1 * time.Minute))
	if err != nil {
		return video, fmt.Errorf("could not generate presigned url: %w", err)
	}

	signedVideo := video
	signedVideo.VideoURL = &presignedURL

	return signedVideo, nil
}*/
