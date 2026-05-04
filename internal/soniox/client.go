package soniox

import (
	"fmt"
	"os"
)

const baseURL = "https://api.soniox.com/v1"

type Client struct {
	apiKey string
}

func NewClient() *Client {
	return &Client{apiKey: os.Getenv("SONIOX_API_KEY")}
}

type TranscriptionRequest struct {
	Model                        string `json:"model"`
	AudioURL                     string `json:"audio_url,omitempty"`
	FileID                       string `json:"file_id,omitempty"`
	EnableSpeakerDiarization     bool   `json:"enable_speaker_diarization"`
	EnableLanguageIdentification bool   `json:"enable_language_identification"`
}

type TranscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type Transcript struct {
	Text   string  `json:"text"`
	Tokens []Token `json:"tokens"`
}

type Token struct {
	Text    string `json:"text"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
}

func (c *Client) CreateTranscription(audioPath string) (string, error) {
	// For simplicity, first upload file or use public URL.
	// Test e public URL use koro or file upload add koro later.
	return "", fmt.Errorf("implement file upload or use public URL")
}

// Better for test: public URL use koro or implement file upload
