package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/snake/internal/audio"
	"github.com/ztkent/snake/internal/highscores"
)

// NewGame creates and initializes a new game instance
func NewGame(screenWidth, screenHeight int32) *Game {
	scores, err := highscores.LoadHighScores()
	if err != nil {
		scores = make([]highscores.HighScore, 0)
	}

	am := audio.NewAudioManager()
	am.LoadResources()

	game := &Game{
		state:        StateMainMenu,
		volume:       100,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		running:      true,
		menu:         NewMenuState(screenWidth, screenHeight),
		highScores:   scores,
		audio:        am,
	}
	return game
}

// Run is the main game loop
func (g *Game) Run() {
	for g.running && !rl.WindowShouldClose() {
		switch g.state {
		case StateMainMenu:
			g.running = g.openMainMenu()
		case StateSettings:
			g.openSettingsMenu()
		case StateGame:
			g.StartGame()
		case StateGameOver:
			g.openGameOverScreen()
		case StateHighScores:
			g.openHighScoresScreen()
		}
	}
}

func main() {
	screenWidth := int32(800)
	screenHeight := int32(450)
	rl.InitWindow(screenWidth, screenHeight, "snake v0")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	game := NewGame(screenWidth, screenHeight)
	defer game.audio.UnloadResources()
	defer rl.UnloadFont(game.menu.font)
	game.Run()
}
