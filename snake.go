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

	food := Food{size: gridSize}
	g.spawnFood(&food, snake.segments)

	lastUpdateTime := float32(0)
	pauseStartTime := float32(0)
	totalPauseTime := float32(0)

	for {
		if rl.IsKeyPressed(rl.KeyEscape) {
			g.state = StatePaused
			pauseStartTime = float32(rl.GetTime())
			if !g.openPauseScreen() {
				return // Exit to main menu if openPauseScreen returns false
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

		if deltaTime >= 1.0/15.0 {
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

			// Check food collision
			if g.checkFoodCollision(newHead, snake.size, food) {
				// Increment score
				g.score.points++
				// Grow snake
				snake.segments = append([]rl.Vector2{newHead}, snake.segments...)
				g.spawnFood(&food, snake.segments)
			} else {
				// Move snake
				snake.segments = append([]rl.Vector2{newHead}, snake.segments[:len(snake.segments)-1]...)
			}

			lastUpdateTime = float32(currentTime)
		}

		// Update duration (subtracting total pause time)
		g.score.duration = float32(rl.GetTime()) - g.score.startTime - totalPauseTime

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

		// Draw food
		rl.DrawRectangleV(food.position, rl.Vector2{X: food.size, Y: food.size}, rl.Red)

		// Draw snake
		g.drawSnake(snake)

		rl.EndDrawing()
	}
}
