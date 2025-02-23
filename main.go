package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// GameState represents the current state of the game
type GameState int

const (
	StateMainMenu GameState = iota
	StateGame
	StateSettings
	StateGameOver
)

// Game handles the game state and settings
type Game struct {
	state          GameState
	volume         float32
	screenWidth    int32
	screenHeight   int32
	running        bool
	buttonReleased bool
	font           rl.Font
}

// NewGame creates and initializes a new game instance
func NewGame(screenWidth, screenHeight int32) *Game {
	game := &Game{
		state:          StateMainMenu,
		volume:         100,
		screenWidth:    screenWidth,
		screenHeight:   screenHeight,
		running:        true,
		buttonReleased: true,
	}

	// Load custom font
	game.font = rl.LoadFont("assets/RetroGaming.ttf")
	return game
}

// Helper method to handle button clicks safely
func (g *Game) handleButtonClick() bool {
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if g.buttonReleased {
			g.buttonReleased = false
			return true
		}
	} else {
		g.buttonReleased = true
	}
	return false
}

// Run is the main game loop
func (g *Game) Run() {
	for g.running {
		switch g.state {
		case StateMainMenu:
			g.running = g.openMainMenu()
		case StateSettings:
			g.openSettingsMenu()
		case StateGame:
			// TODO: Implement game state
			g.state = StateMainMenu
		case StateGameOver:
			// TODO: Implement game over state
			g.state = StateMainMenu
		}
	}
}

type Button struct {
	rect     rl.Rectangle
	text     string
	fontSize int32
	color    rl.Color
	font     rl.Font // Add font field
}

func NewButton(x, y, width, height float32, text string, fontSize int32, font rl.Font) Button {
	return Button{
		rect:     rl.NewRectangle(x, y, width, height),
		text:     text,
		fontSize: fontSize,
		color:    rl.LightGray,
		font:     font,
	}
}

func (b *Button) Draw() {
	rl.DrawRectangleRec(b.rect, b.color)
	textSize := rl.MeasureTextEx(b.font, b.text, float32(b.fontSize), 1)
	rl.DrawTextEx(
		b.font,
		b.text,
		rl.Vector2{
			X: b.rect.X + (b.rect.Width-textSize.X)/2,
			Y: b.rect.Y + (b.rect.Height-textSize.Y)/2,
		},
		float32(b.fontSize),
		1,
		rl.DarkGray,
	)
}

func (b *Button) IsHovered(mousePoint rl.Vector2) bool {
	return rl.CheckCollisionPointRec(mousePoint, b.rect)
}

func main() {
	screenWidth := int32(800)
	screenHeight := int32(450)
	rl.InitWindow(screenWidth, screenHeight, "snake v0")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	game := NewGame(screenWidth, screenHeight)
	defer rl.UnloadFont(game.font)
	game.Run()
}

// openMainMenu displays the main menu interface with Start, Settings, and Exit buttons.
func (g *Game) openMainMenu() bool {
	buttonWidth := float32(200)
	buttonHeight := float32(50)
	buttonSpacing := float32(20)
	startY := float32(g.screenHeight)/2 - (buttonHeight*3+buttonSpacing*2)/2

	startButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY,
		buttonWidth,
		buttonHeight,
		"Start",
		30,
		g.font,
	)
	settingsButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+buttonHeight+buttonSpacing,
		buttonWidth,
		buttonHeight,
		"Settings",
		30,
		g.font,
	)
	exitButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+2*(buttonHeight+buttonSpacing),
		buttonWidth,
		buttonHeight,
		"Exit",
		30,
		g.font,
	)

	// Add title configuration
	titleText := "SNAKE!"
	titleFontSize := float32(80)
	titleSize := rl.MeasureTextEx(g.font, titleText, titleFontSize, 1)
	titleY := startY - titleSize.Y - buttonSpacing

	for !rl.WindowShouldClose() {
		mousePoint := rl.GetMousePosition()

		// Update button states
		if startButton.IsHovered(mousePoint) {
			startButton.color = rl.Gray
			if g.handleButtonClick() {
				g.state = StateGame
				return true
			}
		} else {
			startButton.color = rl.LightGray
		}

		if settingsButton.IsHovered(mousePoint) {
			settingsButton.color = rl.Gray
			if g.handleButtonClick() {
				g.state = StateSettings
				return true
			}
		} else {
			settingsButton.color = rl.LightGray
		}

		if exitButton.IsHovered(mousePoint) {
			exitButton.color = rl.Gray
			if g.handleButtonClick() {
				return false
			}
		} else {
			exitButton.color = rl.LightGray
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Draw title with custom font
		rl.DrawTextEx(
			g.font,
			titleText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - titleSize.X/2,
				Y: titleY,
			},
			titleFontSize,
			1,
			rl.DarkGreen,
		)

		startButton.Draw()
		settingsButton.Draw()
		exitButton.Draw()

		rl.EndDrawing()
	}
	return false
}

// openSettingsMenu displays the settings interface with volume control and a back button.
func (g *Game) openSettingsMenu() {
	buttonWidth := float32(200)
	buttonHeight := float32(50)
	buttonSpacing := float32(20)
	startY := float32(g.screenHeight)/2 - (buttonHeight*2+buttonSpacing)/2

	volumeText := fmt.Sprintf("Volume: %0.f%%", g.volume)

	volumeButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY,
		buttonWidth,
		buttonHeight,
		volumeText,
		30,
		g.font,
	)

	backButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+buttonHeight+buttonSpacing,
		buttonWidth,
		buttonHeight,
		"Back",
		30,
		g.font,
	)

	for {
		// Escape to return to main menu
		if rl.IsKeyReleased(rl.KeyEscape) {
			g.state = StateMainMenu
			return
		}

		mousePoint := rl.GetMousePosition()

		// Handle volume control
		if volumeButton.IsHovered(mousePoint) {
			volumeButton.color = rl.Gray
			if rl.IsKeyDown(rl.KeyLeft) {
				vol := float32(max(0, float64(g.volume-1)))
				if vol < 0 {
					vol = 0
				}
				g.volume = vol
				volumeText = fmt.Sprintf("Volume: %0.f%%", g.volume)
				volumeButton.text = volumeText
			}
			if rl.IsKeyDown(rl.KeyRight) {
				vol := float32(min(100, float64(g.volume+1)))
				if vol > 100 {
					vol = 100
				}
				g.volume = vol
				volumeText = fmt.Sprintf("Volume: %0.f%%", g.volume)
				volumeButton.text = volumeText
			}
		} else {
			volumeButton.color = rl.LightGray
		}

		// Handle back button
		if backButton.IsHovered(mousePoint) {
			backButton.color = rl.Gray
			if g.handleButtonClick() {
				g.state = StateMainMenu
				return
			}
		} else {
			backButton.color = rl.LightGray
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		volumeButton.Draw()
		backButton.Draw()

		// Draw instructions
		instructionsText := "Use Left/Right arrows to adjust volume"
		fontSize := float32(20)
		textSize := rl.MeasureTextEx(g.font, instructionsText, fontSize, 1)
		rl.DrawTextEx(
			g.font,
			instructionsText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - textSize.X/2,
				Y: startY - buttonSpacing*2,
			},
			fontSize,
			1,
			rl.DarkGray,
		)

		rl.EndDrawing()
	}
}

// Helper functions for min/max operations
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
