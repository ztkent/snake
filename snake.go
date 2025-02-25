package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/ztkent/snake/internal/audio"
	"github.com/ztkent/snake/internal/highscores"
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

const (
	gridSize     = 20  // Size of each grid cell
	initialSpeed = 200 // Pixels per second
)

type Direction struct {
	X float32
	Y float32
}

type GameSnake struct {
	segments  []rl.Vector2
	direction Direction
	speed     float32
	size      float32
}

type Food struct {
	position rl.Vector2
	size     float32
}
type Bomb struct {
	position rl.Vector2
	size     float32
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
	highScores   []highscores.HighScore
	audio        *audio.AudioManager
}

type Score struct {
	points    int
	duration  float32
	startTime float32
}

// StartGame implements the main game loop for snake game:
//
// Initialization:
// - Resets score and starts tracking game duration
// - Creates initial snake with 2 segments in center screen
// - Spawns first food piece in random valid location
//
// Main Loop Components:
//
// Input Handling:
// - Window close (X) detection for game exit
// - Arrow key detection for snake direction changes
// - Prevents 180° turns by checking opposite direction
//
// Game State Updates (15 FPS lock):
// - Calculates new head position based on current direction
// - Handles screen wrapping when snake crosses borders
// - Checks for collisions with:
//   - Snake's own body (game over condition)
//   - Food (triggers growth and score increment)
//
// - Updates snake movement:
//   - Adds new head segment
//   - Either removes tail (normal movement)
//   - Or keeps tail (when growing from food)
//
// Time Management:
// - Tracks total game duration
// - Maintains consistent game speed (15 FPS)
// - Adjusts for any pause time
//
// Rendering (60 FPS):
// - Clears screen with dark gray background
// - Draws current score in top right
// - Shows game duration below score
// - Renders food as red square
// - Draws snake with:
//   - Green body segments
//   - Dark green head
//   - White eye (position based on direction)
//
// Loop Exit Conditions:
// - Player closes window (returns to main menu)
// - Snake collides with itself (triggers game over screen)
func (g *Game) StartGame() {
	// Start the game music
	g.audio.SetVolume(g.volume)
	g.audio.PlayMusic(&g.audio.GameMusic)

	// Initialize score
	g.score = Score{
		points:    0,
		startTime: float32(rl.GetTime()),
		duration:  0,
	}

	// Initialize snake in the middle of the screen
	snake := GameSnake{
		segments: []rl.Vector2{
			{X: float32(g.screenWidth / 2), Y: float32(g.screenHeight / 2)},
			{X: float32(g.screenWidth/2) - gridSize, Y: float32(g.screenHeight / 2)},
		},
		direction: Direction{X: 1, Y: 0},
		speed:     initialSpeed,
		size:      gridSize,
	}

	foods := make([]Food, 0)
	bombs := make([]Bomb, 0)
	lastUpdateTime := float32(0)
	pauseStartTime := float32(0)
	totalPauseTime := float32(0)

	for {
		// Update music at consistent intervals
		currentTime := rl.GetTime()
		deltaTime := float32(currentTime) - lastUpdateTime
		if deltaTime >= 1.0/60.0 { // Update at 60Hz
			g.audio.UpdateMusic()
		}

		if rl.IsKeyPressed(rl.KeyEscape) {
			g.state = StatePaused
			pauseStartTime = float32(rl.GetTime())
			if !g.openPauseScreen() {
				return // Exit to main menu if 'exit' is selected
			}
			// Calculate pause duration and adjust times
			totalPauseTime += float32(rl.GetTime()) - pauseStartTime
			lastUpdateTime = float32(rl.GetTime())
			continue
		} else if rl.WindowShouldClose() {
			g.state = StateMainMenu
			g.running = false
			return
		}

		// Handle input
		if rl.IsKeyPressed(rl.KeyUp) && snake.direction.Y != 1 {
			snake.direction = Direction{X: 0, Y: -1}
		}
		if rl.IsKeyPressed(rl.KeyDown) && snake.direction.Y != -1 {
			snake.direction = Direction{X: 0, Y: 1}
		}
		if rl.IsKeyPressed(rl.KeyLeft) && snake.direction.X != 1 {
			snake.direction = Direction{X: -1, Y: 0}
		}
		if rl.IsKeyPressed(rl.KeyRight) && snake.direction.X != -1 {
			snake.direction = Direction{X: 1, Y: 0}
		}

		currentTime = rl.GetTime()
		deltaTime = float32(currentTime) - lastUpdateTime

		if deltaTime >= 1.0/15.0 { // 15 FPS lock
			// Update snake position
			newHead := rl.Vector2{
				X: snake.segments[0].X + snake.direction.X*snake.size,
				Y: snake.segments[0].Y + snake.direction.Y*snake.size,
			}

			// Handle screen wrapping
			newHead = g.wrapPosition(newHead, snake.size)

			// Check self-collision
			if g.checkSelfCollision(newHead, snake.segments) {
				g.audio.PlaySound(&g.audio.GameOverSFX)
				g.state = StateGameOver
				g.audio.PlayMusic(&g.audio.MenuMusic)
				return
			}

			// Check bomb collision with all bombs
			for _, bomb := range bombs {
				if g.checkBombCollision(newHead, snake.size, bomb) {
					g.audio.PlaySound(&g.audio.GameOverSFX)
					g.state = StateGameOver
					g.audio.PlayMusic(&g.audio.MenuMusic)
					return
				}
			}

			// Check food collision with all food pieces
			eaten := -1
			for i, food := range foods {
				if g.checkFoodCollision(newHead, snake.size, food) {
					g.score.points++
					g.audio.PlaySound(&g.audio.CollectSFX)
					snake.segments = append([]rl.Vector2{newHead}, snake.segments...)
					eaten = i
					break
				}
			}

			// Remove eaten food
			if eaten >= 0 {
				foods = append(foods[:eaten], foods[eaten+1:]...)
			}

			// Spawn new food if none exists
			if len(foods) == 0 {
				currentGameTime := float32(rl.GetTime()) - g.score.startTime - totalPauseTime
				g.spawnFoodAndBombs(&foods, &bombs, snake.segments, currentGameTime)
			} else {
				// Move snake
				snake.segments = append([]rl.Vector2{newHead}, snake.segments[:len(snake.segments)-1]...)
			}

			lastUpdateTime = float32(currentTime)

			// Update duration (subtracting total pause time)
			g.score.duration = float32(rl.GetTime()) - g.score.startTime - totalPauseTime
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.DarkGray)

		// Draw score
		scoreText := fmt.Sprintf("Score: %d", g.score.points)
		durationText := fmt.Sprintf("Time: %.1fs", g.score.duration)
		fontSize := float32(20)

		// Draw score
		scoreSize := rl.MeasureTextEx(g.menu.font, scoreText, fontSize, 1)
		rl.DrawTextEx(
			g.menu.font,
			scoreText,
			rl.Vector2{
				X: float32(g.screenWidth) - scoreSize.X - 10,
				Y: 10,
			},
			fontSize,
			1,
			rl.White,
		)

		// Draw duration below score
		durationSize := rl.MeasureTextEx(g.menu.font, durationText, fontSize, 1)
		rl.DrawTextEx(
			g.menu.font,
			durationText,
			rl.Vector2{
				X: float32(g.screenWidth) - durationSize.X - 10,
				Y: scoreSize.Y + 15,
			},
			fontSize,
			1,
			rl.White,
		)

		// Draw all food pieces
		for _, food := range foods {
			rl.DrawRectangleV(food.position, rl.Vector2{X: food.size, Y: food.size}, rl.Gold)
		}

		// Draw all bombs
		for _, bomb := range bombs {
			rl.DrawRectangleV(bomb.position, rl.Vector2{X: bomb.size, Y: bomb.size}, rl.Red)
		}

		// Draw snake
		g.drawSnake(snake)
		rl.EndDrawing()
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

func (g *Game) checkBombCollision(head rl.Vector2, size float32, bomb Bomb) bool {
	return rl.CheckCollisionRecs(
		rl.NewRectangle(head.X, head.Y, size, size),
		rl.NewRectangle(bomb.position.X, bomb.position.Y, bomb.size, bomb.size),
	)
}
func (g *Game) drawSnake(snake GameSnake) {
	for i, segment := range snake.segments {
		if i == 0 {
			// Draw head
			rl.DrawRectangleV(segment, rl.Vector2{X: snake.size, Y: snake.size}, rl.DarkGreen)
		} else {
			// Draw body segments
			rl.DrawRectangleV(segment, rl.Vector2{X: snake.size, Y: snake.size}, rl.Green)
		}
	}
}

func (g *Game) spawnFoodAndBombs(foods *[]Food, bombs *[]Bomb, snakeSegments []rl.Vector2, currentGameTime float32) {
	gridWidth := g.screenWidth / int32(gridSize)
	gridHeight := g.screenHeight / int32(gridSize)

	// Calculate food and bomb counts
	foodCount := int(currentGameTime/10) + 1
	if foodCount > 6 {
		foodCount = 6
	}

	bombCount := 0
	if foodCount > 1 {
		bombCount = foodCount / 2
	}

	// Create array to track occupied positions
	occupied := make(map[string]bool)
	for _, segment := range snakeSegments {
		key := fmt.Sprintf("%d,%d", int(segment.X), int(segment.Y))
		occupied[key] = true
	}

	// Clear existing food and bombs
	*foods = make([]Food, 0, foodCount)
	*bombs = make([]Bomb, 0, bombCount)

	// Spawn food first
	for len(*foods) < foodCount {
		x := float32(rl.GetRandomValue(0, gridWidth-1)) * gridSize
		y := float32(rl.GetRandomValue(0, gridHeight-1)) * gridSize

		key := fmt.Sprintf("%d,%d", int(x), int(y))
		if !occupied[key] {
			*foods = append(*foods, Food{
				position: rl.Vector2{X: x, Y: y},
				size:     gridSize,
			})
			occupied[key] = true

			// Mark adjacent cells as occupied for bomb spacing
			for dx := -1; dx <= 1; dx++ {
				for dy := -1; dy <= 1; dy++ {
					nx := int(x) + dx*int(gridSize)
					ny := int(y) + dy*int(gridSize)
					adjKey := fmt.Sprintf("%d,%d", nx, ny)
					occupied[adjKey] = true
				}
			}
		}
	}

	// Then spawn bombs
	if bombCount > 0 {
		for len(*bombs) < bombCount {
			x := float32(rl.GetRandomValue(0, gridWidth-1)) * gridSize
			y := float32(rl.GetRandomValue(0, gridHeight-1)) * gridSize

			key := fmt.Sprintf("%d,%d", int(x), int(y))
			if !occupied[key] {
				*bombs = append(*bombs, Bomb{
					position: rl.Vector2{X: x, Y: y},
					size:     gridSize,
				})
				occupied[key] = true
			}
		}
	}
}
