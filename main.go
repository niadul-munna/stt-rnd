package main

import (
	"fmt"
	"log"
	"os"

	"vidinfra-stt-test/internal/ffmpeg"
	"vidinfra-stt-test/internal/processor"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	videoPath := "test_video.mp4"

	fmt.Println("=== Vidinfra Auto Subtitle Test Starting ===")

	if os.Getenv("SONIOX_API_KEY") == "" {
		log.Fatal("❌ SONIOX_API_KEY error on .env file")
	}
	fmt.Println("✅ Soniox API Key Loaded")

	// Step 1: Audio extract
	audioPath, err := ffmpeg.ExtractAudio(videoPath)
	if err != nil {
		log.Fatalf("Audio extract failed: %v", err)
	}
	defer os.Remove(audioPath)

	fmt.Println("✅ Audio extracted successfully:", audioPath)

	// Step 2: Soniox processing
	result, err := processor.ProcessWithSoniox(audioPath)
	if err != nil {
		log.Fatalf("Soniox processing failed: %v", err)
	}

	fmt.Println("\n✅ Transcription Completed!")
	fmt.Println("Text length:", len(result.Text))

	// Safe text print
	if len(result.Text) > 0 {
		sampleLen := 200
		if len(result.Text) < sampleLen {
			sampleLen = len(result.Text)
		}
		fmt.Println("\nSample Text:\n", result.Text[:sampleLen])
	}

	if len(result.SRT) > 0 {
		fmt.Println("\nSRT Preview (first 300 chars):")
		srtLen := 300
		if len(result.SRT) < srtLen {
			srtLen = len(result.SRT)
		}
		fmt.Println(result.SRT[:srtLen])
	}
}
