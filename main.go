package main

import (
	"fmt"
	"math"
	"time"

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

// Sprite represents a falling pixel element in the background
type Sprite struct {
	position rl.Vector2
	speed    float32
	size     float32
	color    rl.Color
}

// TurnPoint represents a point where the snake changes direction
type TurnPoint struct {
	position  rl.Vector2
	direction float32
}

// SnakeSegment represents a segment of the snake
type SnakeSegment struct {
	position  rl.Vector2
	direction float32
}

// MenuState handles menu-specific UI elements and animations
type MenuState struct {
	sprites        []Sprite
	snakePos       rl.Vector2
	snakeDir       float32
	snakeSpeed     float32
	snakeSize      float32
	snakeLength    int
	snakeSegments  []SnakeSegment
	turnPoints     []TurnPoint
	font           rl.Font
	buttonReleased bool
	screenWidth    int32
	screenHeight   int32
}

// Game handles core game state
type Game struct {
	state        GameState
	volume       float32
	screenWidth  int32
	screenHeight int32
	running      bool
	menu         *MenuState
	score        Score
	highScores   []HighScore
}

type Score struct {
	points    int
	duration  float32
	startTime float32
}

// NewGame creates and initializes a new game instance
func NewGame(screenWidth, screenHeight int32) *Game {
	menu := &MenuState{
		sprites:        make([]Sprite, 50),
		snakePos:       rl.Vector2{X: 0, Y: float32(screenHeight - 40)},
		snakeDir:       1,
		snakeSpeed:     200,
		snakeSize:      10,
		snakeLength:    5,
		snakeSegments:  make([]SnakeSegment, 12),
		turnPoints:     make([]TurnPoint, 0),
		buttonReleased: true,
		screenWidth:    screenWidth, // Initialize screen dimensions
		screenHeight:   screenHeight,
	}

	// Initialize menu elements
	for i := range menu.sprites {
		menu.sprites[i] = newRandomSprite(screenWidth)
	}

	// Initialize snake segments with position and direction
	for i := 0; i < menu.snakeLength; i++ {
		menu.snakeSegments[i] = SnakeSegment{
			position: rl.Vector2{
				X: menu.snakePos.X - float32(i)*menu.snakeSize*1.2,
				Y: menu.snakePos.Y,
			},
			direction: 1,
		}
	}

	menu.font = rl.LoadFont("assets/RetroGaming.ttf")

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
		menu:         menu,
		highScores:   scores,
	}
}

// Create a new random sprite
func newRandomSprite(screenWidth int32) Sprite {
	return Sprite{
		position: rl.Vector2{
			X: float32(rl.GetRandomValue(0, screenWidth)),
			Y: float32(rl.GetRandomValue(-100, 0)),
		},
		speed: float32(rl.GetRandomValue(100, 200)) / 100.0,
		size:  float32(rl.GetRandomValue(2, 6)),
		color: rl.Color{
			R: uint8(rl.GetRandomValue(0, 100)),
			G: uint8(rl.GetRandomValue(100, 255)),
			B: uint8(rl.GetRandomValue(0, 100)),
			A: uint8(rl.GetRandomValue(100, 200)),
		},
	}
}

// Update and draw background sprites
func (m *MenuState) updateBackground() {
	deltaTime := rl.GetFrameTime()

	for i := range m.sprites {
		// Update position
		m.sprites[i].position.Y += m.sprites[i].speed * deltaTime * 100

		// Reset sprite if it's out of screen
		if m.sprites[i].position.Y > float32(m.screenHeight) {
			m.sprites[i] = newRandomSprite(m.screenWidth)
		}

		// Draw sprite
		rl.DrawRectangleV(
			m.sprites[i].position,
			rl.Vector2{X: m.sprites[i].size, Y: m.sprites[i].size},
			m.sprites[i].color,
		)
	}
}

// Helper method to handle button clicks safely
func (m *MenuState) handleButtonClick() bool {
	if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if m.buttonReleased {
			m.buttonReleased = false
			return true
		}
	} else {
		m.buttonReleased = true
	}
	return false
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

type Button struct {
	rect     rl.Rectangle
	text     string
	fontSize int32
	color    rl.Color
	font     rl.Font
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
	defer rl.UnloadFont(game.menu.font)
	game.Run()
}

// openMainMenu displays the main menu interface with Start, Settings, and Exit buttons.
func (g *Game) openMainMenu() bool {
	buttonWidth := float32(200)
	buttonHeight := float32(50)
	buttonSpacing := float32(20)
	startY := float32(g.screenHeight)/2 - (buttonHeight*4+buttonSpacing*3)/2 // Adjusted for new button

	startButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY,
		buttonWidth,
		buttonHeight,
		"Start",
		30,
		g.menu.font,
	)

	highScoresButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+buttonHeight+buttonSpacing,
		buttonWidth,
		buttonHeight,
		"High Scores",
		30,
		g.menu.font,
	)

	settingsButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+2*(buttonHeight+buttonSpacing),
		buttonWidth,
		buttonHeight,
		"Settings",
		30,
		g.menu.font,
	)

	exitButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+3*(buttonHeight+buttonSpacing),
		buttonWidth,
		buttonHeight,
		"Exit",
		30,
		g.menu.font,
	)

	// Title configuration
	titleText := "SNAKE!"
	titleFontSize := float32(80)
	titleSize := rl.MeasureTextEx(g.menu.font, titleText, titleFontSize, 1)
	titleY := startY - titleSize.Y - buttonSpacing

	for !rl.WindowShouldClose() {
		// Update snake animation
		g.menu.updateSnake()

		mousePoint := rl.GetMousePosition()

		// Update button states
		if startButton.IsHovered(mousePoint) {
			startButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateGame
				return true
			}
		} else {
			startButton.color = rl.LightGray
		}

		if highScoresButton.IsHovered(mousePoint) {
			highScoresButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateHighScores
				return true
			}
		} else {
			highScoresButton.color = rl.LightGray
		}

		if settingsButton.IsHovered(mousePoint) {
			settingsButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateSettings
				return true
			}
		} else {
			settingsButton.color = rl.LightGray
		}

		if exitButton.IsHovered(mousePoint) {
			exitButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				return false
			}
		} else {
			exitButton.color = rl.LightGray
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Draw background first
		g.menu.updateBackground()

		// Draw title with custom font
		rl.DrawTextEx(
			g.menu.font,
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
		highScoresButton.Draw()
		settingsButton.Draw()
		exitButton.Draw()

		// Draw snake at the bottom
		g.menu.drawSnake()

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
		g.menu.font,
	)

	backButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		startY+buttonHeight+buttonSpacing,
		buttonWidth,
		buttonHeight,
		"Back",
		30,
		g.menu.font,
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
			if g.menu.handleButtonClick() {
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
		textSize := rl.MeasureTextEx(g.menu.font, instructionsText, fontSize, 1)
		rl.DrawTextEx(
			g.menu.font,
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
func minInt32(a, b int32) int32 {
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

func (m *MenuState) updateSnake() {
	deltaTime := rl.GetFrameTime()

	// Update head position
	m.snakePos.X += m.snakeSpeed * m.snakeDir * deltaTime

	// Check for wall collisions
	if m.snakePos.X > float32(m.screenWidth)-m.snakeSize {
		m.snakePos.X = float32(m.screenWidth) - m.snakeSize
		m.snakeDir = -1
	} else if m.snakePos.X < 0 {
		m.snakePos.X = 0
		m.snakeDir = 1
	}

	// Update head segment
	m.snakeSegments[0].position = m.snakePos
	m.snakeSegments[0].direction = m.snakeDir

	// Update body segments
	spacing := m.snakeSize * 1.2
	for i := 1; i < m.snakeLength; i++ {
		prev := m.snakeSegments[i-1]
		curr := &m.snakeSegments[i]

		// Calculate distance to previous segment
		dist := prev.position.X - curr.position.X
		absDist := float32(math.Abs(float64(dist)))

		// Update position based on current direction
		curr.position.X += m.snakeSpeed * curr.direction * deltaTime

		// Check if segment needs to turn
		if (curr.direction > 0 && curr.position.X >= prev.position.X-spacing) ||
			(curr.direction < 0 && curr.position.X <= prev.position.X+spacing) {
			// Maintain spacing from previous segment
			if curr.direction > 0 {
				curr.position.X = prev.position.X - spacing
			} else {
				curr.position.X = prev.position.X + spacing
			}

			// Only change direction when properly spaced
			if absDist <= spacing*1.1 {
				curr.direction = prev.direction
			}
		}

		curr.position.Y = m.snakePos.Y
	}
}

func (m *MenuState) drawSnake() {
	// Draw body segments first
	for i := m.snakeLength - 1; i > 0; i-- {
		segment := m.snakeSegments[i]
		rl.DrawRectangleV(
			segment.position,
			rl.Vector2{X: m.snakeSize, Y: m.snakeSize},
			rl.Green,
		)
	}

	// Draw head
	headColor := rl.DarkGreen
	if m.snakeDir > 0 {
		// Draw eyes on right side when moving right
		rl.DrawRectangleV(m.snakePos, rl.Vector2{X: m.snakeSize, Y: m.snakeSize}, headColor)
		rl.DrawCircleV(rl.Vector2{X: m.snakePos.X + m.snakeSize*0.7, Y: m.snakePos.Y + m.snakeSize*0.3}, 2, rl.White)
	} else {
		// Draw eyes on left side when moving left
		rl.DrawRectangleV(m.snakePos, rl.Vector2{X: m.snakeSize, Y: m.snakeSize}, headColor)
		rl.DrawCircleV(rl.Vector2{X: m.snakePos.X + m.snakeSize*0.3, Y: m.snakePos.Y + m.snakeSize*0.3}, 2, rl.White)
	}
}

// Display a pause screen with resume and quit buttons
func (g *Game) openPauseScreen() bool {
	buttonWidth := float32(200)
	buttonHeight := float32(50)
	buttonSpacing := float32(20)

	// Create buttons
	resumeButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		float32(g.screenHeight)*0.6,
		buttonWidth,
		buttonHeight,
		"Resume",
		30,
		g.menu.font,
	)

	quitButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		float32(g.screenHeight)*0.6+buttonHeight+buttonSpacing,
		buttonWidth,
		buttonHeight,
		"Quit to Menu",
		30,
		g.menu.font,
	)

	// Text configuration
	pauseText := "PAUSED"
	titleFontSize := float32(60)
	statsFontSize := float32(30)
	titleSize := rl.MeasureTextEx(g.menu.font, pauseText, titleFontSize, 1)

	for {
		mousePoint := rl.GetMousePosition()

		// Handle button states
		if resumeButton.IsHovered(mousePoint) {
			resumeButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateGame
				return true
			}
		} else {
			resumeButton.color = rl.LightGray
		}

		if quitButton.IsHovered(mousePoint) {
			quitButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateMainMenu
				return false
			}
		} else {
			quitButton.color = rl.LightGray
		}

		rl.BeginDrawing()
		// Draw semi-transparent overlay
		rl.DrawRectangle(0, 0, g.screenWidth, g.screenHeight, rl.Color{R: 0, G: 0, B: 0, A: 120})

		// Draw pause text
		rl.DrawTextEx(
			g.menu.font,
			pauseText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - titleSize.X/2,
				Y: float32(g.screenHeight) * 0.2,
			},
			titleFontSize,
			1,
			rl.White,
		)

		// Draw score
		scoreText := fmt.Sprintf("Score: %d", g.score.points)
		timeText := fmt.Sprintf("Time: %.1fs", g.score.duration)

		scoreSize := rl.MeasureTextEx(g.menu.font, scoreText, statsFontSize, 1)
		rl.DrawTextEx(
			g.menu.font,
			scoreText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - scoreSize.X/2,
				Y: float32(g.screenHeight) * 0.4,
			},
			statsFontSize,
			1,
			rl.Green,
		)

		// Draw time
		timeSize := rl.MeasureTextEx(g.menu.font, timeText, statsFontSize, 1)
		rl.DrawTextEx(
			g.menu.font,
			timeText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - timeSize.X/2,
				Y: float32(g.screenHeight)*0.4 + scoreSize.Y + buttonSpacing/2,
			},
			statsFontSize,
			1,
			rl.Green,
		)

		// Draw buttons
		resumeButton.Draw()
		quitButton.Draw()

		rl.EndDrawing()

		if rl.IsKeyPressed(rl.KeyEscape) {
			g.state = StateGame
			return true
		}
	}
}

func (g *Game) wrapPosition(pos rl.Vector2, size float32) rl.Vector2 {
	if pos.X >= float32(g.screenWidth) {
		pos.X = 0
	} else if pos.X < 0 {
		pos.X = float32(g.screenWidth) - size
	}
	if pos.Y >= float32(g.screenHeight) {
		pos.Y = 0
	} else if pos.Y < 0 {
		pos.Y = float32(g.screenHeight) - size
	}
	return pos
}

func (g *Game) checkSelfCollision(head rl.Vector2, segments []rl.Vector2) bool {
	for i := 1; i < len(segments); i++ {
		if head.X == segments[i].X && head.Y == segments[i].Y {
			return true
		}
	}
	return false
}

func (g *Game) checkFoodCollision(head rl.Vector2, size float32, food Food) bool {
	return rl.CheckCollisionRecs(
		rl.NewRectangle(head.X, head.Y, size, size),
		rl.NewRectangle(food.position.X, food.position.Y, food.size, food.size),
	)
}

func (g *Game) drawSnake(snake GameSnake) {
	for i, segment := range snake.segments {
		color := rl.Green
		if i == 0 {
			color = rl.DarkGreen // Head color
			// Draw eyes based on direction
			eyeOffset := float32(0.3)
			if snake.direction.X > 0 {
				eyeOffset = 0.7
			}
			rl.DrawCircleV(
				rl.Vector2{
					X: segment.X + snake.size*eyeOffset,
					Y: segment.Y + snake.size*0.3,
				},
				2,
				rl.White,
			)
		}
		rl.DrawRectangleV(segment, rl.Vector2{X: snake.size, Y: snake.size}, color)
	}
}

func (g *Game) spawnFood(food *Food, snakeSegments []rl.Vector2) {
	gridWidth := g.screenWidth / int32(gridSize)
	gridHeight := g.screenHeight / int32(gridSize)

	for {
		x := float32(rl.GetRandomValue(0, gridWidth-1)) * gridSize
		y := float32(rl.GetRandomValue(0, gridHeight-1)) * gridSize

		conflict := false
		for _, segment := range snakeSegments {
			if x == segment.X && y == segment.Y {
				conflict = true
				break
			}
		}

		if !conflict {
			food.position = rl.Vector2{X: x, Y: y}
			break
		}
	}
}

// Game over screen, displays final score and time
func (g *Game) openGameOverScreen() {
	buttonWidth := float32(240)
	buttonHeight := float32(50)
	buttonSpacing := float32(20)

	// Create exit button
	exitButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		float32(g.screenHeight)*0.7,
		buttonWidth,
		buttonHeight,
		"Back to Menu",
		30,
		g.menu.font,
	)

	// Game Over text configuration
	gameOverText := "GAME OVER!"
	titleFontSize := float32(60)
	titleSize := rl.MeasureTextEx(g.menu.font, gameOverText, titleFontSize, 1)

	// Score text configuration
	scoreText := fmt.Sprintf("Final Score: %d", g.score.points)
	timeText := fmt.Sprintf("Time: %.1fs", g.score.duration)
	statsFontSize := float32(30)

	// Check for high score
	isNewHighScore := IsHighScore(g.score.points, g.highScores)
	if isNewHighScore {
		newScore := HighScore{
			Score:    g.score.points,
			Duration: g.score.duration,
			Date:     time.Now().Format("2006-01-02"),
		}
		g.highScores = UpdateHighScores(g.highScores, newScore)
		SaveHighScores(g.highScores)
	}

	// Create high score text
	highScoreText := "NEW HIGH SCORE!"
	highScoreFontSize := float32(28)
	highScoreSize := rl.MeasureTextEx(g.menu.font, highScoreText, highScoreFontSize, 1)

	for {
		mousePoint := rl.GetMousePosition()
		// Handle button interaction
		if exitButton.IsHovered(mousePoint) {
			exitButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateMainMenu
				return
			}
		} else {
			exitButton.color = rl.LightGray
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Draw background
		g.menu.updateBackground()

		// Draw game over text
		rl.DrawTextEx(
			g.menu.font,
			gameOverText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - titleSize.X/2,
				Y: float32(g.screenHeight) * 0.2,
			},
			titleFontSize,
			1,
			rl.Maroon,
		)

		scoreSize := rl.MeasureTextEx(g.menu.font, scoreText, statsFontSize, 1)

		// Draw high score notification if applicable
		if isNewHighScore {
			rl.DrawTextEx(
				g.menu.font,
				highScoreText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - highScoreSize.X/2,
					Y: float32(g.screenHeight) * 0.35,
				},
				highScoreFontSize,
				1,
				rl.Gold,
			)
			// Draw score
			rl.DrawTextEx(
				g.menu.font,
				scoreText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - scoreSize.X/2,
					Y: float32(g.screenHeight) * 0.45,
				},
				statsFontSize,
				1,
				rl.DarkGreen,
			)
			// Draw time
			timeSize := rl.MeasureTextEx(g.menu.font, timeText, statsFontSize, 1)
			rl.DrawTextEx(
				g.menu.font,
				timeText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - timeSize.X/2,
					Y: float32(g.screenHeight)*0.45 + scoreSize.Y + buttonSpacing,
				},
				statsFontSize,
				1,
				rl.DarkGreen,
			)
		} else {
			// Draw score
			rl.DrawTextEx(
				g.menu.font,
				scoreText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - scoreSize.X/2,
					Y: float32(g.screenHeight) * 0.40,
				},
				statsFontSize,
				1,
				rl.DarkGreen,
			)

			// Draw time
			timeSize := rl.MeasureTextEx(g.menu.font, timeText, statsFontSize, 1)
			rl.DrawTextEx(
				g.menu.font,
				timeText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - timeSize.X/2,
					Y: float32(g.screenHeight)*0.40 + scoreSize.Y + buttonSpacing,
				},
				statsFontSize,
				1,
				rl.DarkGreen,
			)
		}

		// Draw exit button
		exitButton.Draw()
		rl.EndDrawing()
	}
}

// Add new method for high scores screen
func (g *Game) openHighScoresScreen() {
	buttonWidth := float32(200)
	buttonHeight := float32(50)

	backButton := NewButton(
		float32(g.screenWidth)/2-buttonWidth/2,
		float32(g.screenHeight)*0.8,
		buttonWidth,
		buttonHeight,
		"Back",
		30,
		g.menu.font,
	)

	titleText := "HIGH SCORES"
	titleFontSize := float32(60)
	statsFontSize := float32(30)
	titleSize := rl.MeasureTextEx(g.menu.font, titleText, titleFontSize, 1)

	for {
		if rl.IsKeyReleased(rl.KeyEscape) {
			g.state = StateMainMenu
			return
		}

		mousePoint := rl.GetMousePosition()

		if backButton.IsHovered(mousePoint) {
			backButton.color = rl.Gray
			if g.menu.handleButtonClick() {
				g.state = StateMainMenu
				return
			}
		} else {
			backButton.color = rl.LightGray
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// Draw title
		rl.DrawTextEx(
			g.menu.font,
			titleText,
			rl.Vector2{
				X: float32(g.screenWidth)/2 - titleSize.X/2,
				Y: float32(g.screenHeight) * 0.1,
			},
			titleFontSize,
			1,
			rl.DarkGreen,
		)

		// Draw high scores
		startY := float32(g.screenHeight) * 0.3
		for i, score := range g.highScores {
			scoreText := fmt.Sprintf("%d. Score: %d  Time: %.1fs  (%s)",
				i+1, score.Score, score.Duration, score.Date)
			scoreSize := rl.MeasureTextEx(g.menu.font, scoreText, statsFontSize, 1)
			rl.DrawTextEx(
				g.menu.font,
				scoreText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - scoreSize.X/2,
					Y: startY + float32(i)*statsFontSize*1.5,
				},
				statsFontSize,
				1,
				rl.DarkGray,
			)
		}

		// Draw "No scores yet" if there are no high scores
		if len(g.highScores) == 0 {
			noScoresText := "No scores yet!"
			textSize := rl.MeasureTextEx(g.menu.font, noScoresText, statsFontSize, 1)
			rl.DrawTextEx(
				g.menu.font,
				noScoresText,
				rl.Vector2{
					X: float32(g.screenWidth)/2 - textSize.X/2,
					Y: float32(g.screenHeight) * 0.4,
				},
				statsFontSize,
				1,
				rl.Gray,
			)
		}

		backButton.Draw()
		rl.EndDrawing()
	}
}
