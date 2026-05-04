package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type Result struct {
	Text     string    `json:"text"`
	SRT      string    `json:"srt"`
	Duration string    `json:"duration"`
	Chapters []Chapter `json:"chapters"`
}

type Chapter struct {
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
}

type sonioxFileResponse struct {
	ID string `json:"id"`
}

type sonioxTranscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type sonioxTranscriptResponse struct {
	Text   string  `json:"text"`
	Tokens []Token `json:"tokens"`
}

type Token struct {
	Text    string `json:"text"`
	StartMs int    `json:"start_ms"`
	EndMs   int    `json:"end_ms"`
}

// ProcessWithSoniox - Full Real Implementation
func ProcessWithSoniox(audioPath string) (*Result, error) {
	apiKey := os.Getenv("SONIOX_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SONIOX_API_KEY missing")
	}

	fmt.Println("Uploading audio to Soniox...")

	// 1. Upload File
	fileID, err := uploadFile(audioPath, apiKey)
	if err != nil {
		return nil, err
	}
	fmt.Println("✅ Audio uploaded. File ID:", fileID)

	// 2. Create Transcription
	transID, err := createTranscription(fileID, apiKey)
	if err != nil {
		return nil, err
	}
	fmt.Println("✅ Transcription created. ID:", transID)

	// 3. Poll for result
	transcript, err := pollForTranscription(transID, apiKey)
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ Transcription completed! Text length:", len(transcript.Text))

	srt := generateSRT(transcript.Tokens)
	chapters := generateBasicChapters()

	return &Result{
		Text:     transcript.Text,
		SRT:      srt,
		Duration: "N/A",
		Chapters: chapters,
	}, nil
}

func uploadFile(audioPath, apiKey string) (string, error) {
	file, err := os.Open(audioPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "audio.wav")
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "https://api.soniox.com/v1/files", body)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result sonioxFileResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.ID, nil
}

func createTranscription(fileID, apiKey string) (string, error) {
	payload := map[string]interface{}{
		"model":                      "stt-async-v4", // ← correct latest model
		"file_id":                    fileID,
		"enable_speaker_diarization": true, //if video is Bangla then helpful
		// "language_hints": []string{"bn", "en"},  // if needed then uncomment
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "https://api.soniox.com/v1/transcriptions", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create transcription failed: %s", string(bodyBytes))
	}

	var result sonioxTranscriptionResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return result.ID, nil
}

func pollForTranscription(transID, apiKey string) (*sonioxTranscriptResponse, error) {
	client := &http.Client{}
	statusURL := fmt.Sprintf("https://api.soniox.com/v1/transcriptions/%s", transID)
	transcriptURL := fmt.Sprintf("https://api.soniox.com/v1/transcriptions/%s/transcript", transID)

	fmt.Println("🔄 Polling started... (Dynamic timeout)")

	maxWaitMinutes := 25              // on Production 25-30 min max wait safe (for 2hr video is enough)
	maxAttempts := maxWaitMinutes * 6 // 10 sec interval

	for i := 0; i < maxAttempts; i++ {
		req, _ := http.NewRequest("GET", statusURL, nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var statusResp struct {
			ID           string `json:"id"`
			Status       string `json:"status"`
			Error        string `json:"error,omitempty"`
			ErrorMessage string `json:"error_message,omitempty"`
		}

		json.Unmarshal(bodyBytes, &statusResp)

		elapsedMin := float64(i*10) / 60
		fmt.Printf("Attempt %d | Elapsed: %.1f min | Status: %s\n", i+1, elapsedMin, statusResp.Status)

		if statusResp.Status == "completed" {
			fmt.Println("✅ Transcription completed! Fetching tokens...")

			req, _ = http.NewRequest("GET", transcriptURL, nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err = client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var transcript sonioxTranscriptResponse
			json.NewDecoder(resp.Body).Decode(&transcript)

			fmt.Println("✅ Tokens received:", len(transcript.Tokens))
			return &transcript, nil

		} else if statusResp.Status == "error" {
			return nil, fmt.Errorf("Soniox error: %s", statusResp.ErrorMessage)
		}

		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("transcription timeout after %d minutes", maxWaitMinutes)
}

func generateSRT(tokens []Token) string {
	// Simple SRT generator (improve later)
	srt := "1\n00:00:00,000 --> 00:00:10,000\n"
	if len(tokens) > 0 {
		srt += tokens[0].Text + "\n"
	}
	return srt
}

func generateBasicChapters() []Chapter {
	return []Chapter{
		{Timestamp: "00:00", Title: "started"},
		{Timestamp: "00:30", Title: "main discussion"},
	}
}
