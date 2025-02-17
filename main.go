package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	screenWidth := int32(800)
	screenHeight := int32(450)

	rl.InitWindow(screenWidth, screenHeight, "snake v0")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		text := "Hello World!"
		fontSize := int32(40)
		textWidth := rl.MeasureText(text, fontSize)
		rl.DrawText(
			text,
			screenWidth/2-textWidth/2,
			screenHeight/2-fontSize/2,
			fontSize,
			rl.DarkGray,
		)
		rl.EndDrawing()
	}
}
