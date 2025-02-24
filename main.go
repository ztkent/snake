package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// GameState represents the current state of the game
type GameState int

const (
	StateMainMenu GameState = iota
	StateGame
	StateSettings
	StateGameOver
	StatePaused
	StateHighScores // Add new state
)

// NewGame creates and initializes a new game instance
func NewGame(screenWidth, screenHeight int32) *Game {
	scores, err := LoadHighScores()
	if err != nil {
		scores = make([]HighScore, 0)
	}
	return &Game{
		state:        StateMainMenu,
		volume:       100,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		running:      true,
		menu:         NewMenuState(screenWidth, screenHeight),
		highScores:   scores,
	}
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
	defer rl.UnloadFont(game.menu.font)
	game.Run()
}
