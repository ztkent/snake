package highscores

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
)

const (
	highScoresFile = "highscores.csv"
	maxHighScores  = 3
)

type HighScore struct {
	Score    int
	Duration float32
	Date     string
}

func LoadHighScores() ([]HighScore, error) {
	scores := make([]HighScore, 0)

	// Create file if it doesn't exist
	if _, err := os.Stat(highScoresFile); os.IsNotExist(err) {
		return scores, nil
	}

	file, err := os.Open(highScoresFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if len(record) != 3 {
			continue
		}
		score, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}
		duration, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			continue
		}
		scores = append(scores, HighScore{
			Score:    score,
			Duration: float32(duration),
			Date:     record[2],
		})
	}

	return scores, nil
}

func SaveHighScores(scores []HighScore) error {
	file, err := os.Create(highScoresFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, score := range scores {
		record := []string{
			strconv.Itoa(score.Score),
			fmt.Sprintf("%.1f", score.Duration),
			score.Date,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func IsHighScore(score int, scores []HighScore) bool {
	if len(scores) < maxHighScores {
		return true
	}
	return score > scores[len(scores)-1].Score
}

func UpdateHighScores(scores []HighScore, newScore HighScore) []HighScore {
	scores = append(scores, newScore)
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Duration < scores[j].Duration
		}
		return scores[i].Score > scores[j].Score
	})

	if len(scores) > maxHighScores {
		scores = scores[:maxHighScores]
	}
	return scores
}
