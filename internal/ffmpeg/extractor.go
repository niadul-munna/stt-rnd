package ffmpeg

import (
	"fmt"
	"os/exec"
)

func ExtractAudio(videoPath string) (string, error) {
	audioPath := "temp_audio.wav"

	cmd := exec.Command("ffmpeg", "-i", videoPath,
		"-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1",
		"-y", audioPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %s - %s", err, string(output))
	}
	return audioPath, nil
}
