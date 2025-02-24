package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	lastUpdateTime := float32(0)
	pauseStartTime := float32(0)
	totalPauseTime := float32(0)

	for {
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

		currentTime := rl.GetTime()
		deltaTime := float32(currentTime) - lastUpdateTime

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
				g.state = StateGameOver
				return
			}

			// Check food collision with all food pieces
			eaten := -1
			for i, food := range foods {
				if g.checkFoodCollision(newHead, snake.size, food) {
					g.score.points++
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
				g.spawnFood(&foods, snake.segments, currentGameTime)
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
			rl.DrawRectangleV(food.position, rl.Vector2{X: food.size, Y: food.size}, rl.Red)
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

func (g *Game) spawnFood(foods *[]Food, snakeSegments []rl.Vector2, currentGameTime float32) {
	gridWidth := g.screenWidth / int32(gridSize)
	gridHeight := g.screenHeight / int32(gridSize)

	// Calculate number of food pieces based on time (minimum 1)
	intervals := int(currentGameTime/10) + 1
	if intervals > 5 { // Cap maximum food pieces at 5
		intervals = 5
	}

	// Create array to track occupied positions
	occupied := make(map[string]bool)
	for _, segment := range snakeSegments {
		key := fmt.Sprintf("%d,%d", int(segment.X), int(segment.Y))
		occupied[key] = true
	}

	// Clear existing food
	*foods = make([]Food, 0, intervals)

	// Try to spawn each piece of food
	attempts := 0
	for len(*foods) < intervals && attempts < 100 { // Limit attempts to prevent infinite loop
		x := float32(rl.GetRandomValue(0, gridWidth-1)) * gridSize
		y := float32(rl.GetRandomValue(0, gridHeight-1)) * gridSize

		key := fmt.Sprintf("%d,%d", int(x), int(y))
		if !occupied[key] {
			*foods = append(*foods, Food{
				position: rl.Vector2{X: x, Y: y},
				size:     gridSize,
			})
			occupied[key] = true
		}
		attempts++
	}
}
