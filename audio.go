package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AudioManager struct {
	menuMusic    Music
	gameMusic    Music
	gameOverSFX  Sound
	collectSFX   Sound
	volume       float32
	currentMusic *Music
	isPlaying    bool // Add playing status
}

type Music struct {
	stream rl.Music
	loaded bool
}

type Sound struct {
	sound  rl.Sound
	loaded bool
}

func NewAudioManager() *AudioManager {
	rl.InitAudioDevice()
	return &AudioManager{
		volume: 1.0,
	}
}

func (am *AudioManager) LoadResources() {
	// Load menu music
	menuStream := rl.LoadMusicStream("assets/mainmenu.mp3")
	if !rl.IsMusicValid(menuStream) {
		fmt.Println("Failed to load menu music")
		am.menuMusic = Music{stream: menuStream, loaded: false}
	} else {
		fmt.Println("Menu music loaded successfully")
		am.menuMusic = Music{stream: menuStream, loaded: true}

	}

	// Load game music
	gameStream := rl.LoadMusicStream("assets/gamemusic.mp3")
	if !rl.IsMusicValid(gameStream) {
		fmt.Println("Failed to load game music")
		am.gameMusic = Music{stream: gameStream, loaded: false}
	} else {
		fmt.Println("Game music loaded successfully")
		am.gameMusic = Music{stream: gameStream, loaded: true}
	}

	// Load sound effects
	gameOverSound := rl.LoadSound("assets/gameover.wav")
	am.gameOverSFX = Sound{sound: gameOverSound, loaded: true}

	collectSound := rl.LoadSound("assets/nom.wav")
	rl.SetSoundVolume(collectSound, am.volume*0.5)
	am.collectSFX = Sound{sound: collectSound, loaded: true}

	// Set initial properties
	rl.SetMusicVolume(gameStream, am.volume)
	rl.SetMusicPitch(gameStream, 1.0)
}

func (am *AudioManager) UnloadResources() {
	// Unload music
	if am.menuMusic.loaded {
		rl.UnloadMusicStream(am.menuMusic.stream)
	}
	if am.gameMusic.loaded {
		rl.UnloadMusicStream(am.gameMusic.stream)
	}

	// Unload sound effects
	if am.gameOverSFX.loaded {
		rl.UnloadSound(am.gameOverSFX.sound)
	}
	if am.collectSFX.loaded {
		rl.UnloadSound(am.collectSFX.sound)
	}

	rl.CloseAudioDevice()
}

func (am *AudioManager) PlayMusic(music *Music) {
	if music == nil || !music.loaded {
		fmt.Println("Attempted to play invalid music")
		return
	}

	// Stop current music if playing
	if am.currentMusic != nil && am.currentMusic.loaded {
		fmt.Println("Stopping current music")
		rl.StopMusicStream(am.currentMusic.stream)
		am.isPlaying = false
	}

	am.currentMusic = music
	fmt.Printf("Playing new music (loaded: %v)\n", music.loaded)

	if rl.IsMusicValid(music.stream) {
		rl.SeekMusicStream(music.stream, 0.0)
		rl.PlayMusicStream(music.stream)
		rl.SetMusicVolume(music.stream, am.volume)
		am.isPlaying = true
		fmt.Println("Music started successfully")
	} else {
		fmt.Println("Failed to play music - stream not ready")
	}
}

func (am *AudioManager) UpdateMusic() {
	if am.currentMusic == nil || !am.currentMusic.loaded {
		return
	}

	if !rl.IsMusicStreamPlaying(am.currentMusic.stream) && am.isPlaying {
		fmt.Println("Music ended, restarting...")
		rl.SeekMusicStream(am.currentMusic.stream, 0.0)
		rl.PlayMusicStream(am.currentMusic.stream)
	}

	rl.UpdateMusicStream(am.currentMusic.stream)
}

func (am *AudioManager) PlaySound(sound *Sound) {
	if sound.loaded {
		rl.PlaySound(sound.sound)
	}
}

func (am *AudioManager) SetVolume(volume float32) {
	am.volume = volume / 100.0
	rl.SetMasterVolume(am.volume)
	// Also update current music volume if playing
	if am.currentMusic != nil && am.currentMusic.loaded {
		rl.SetMusicVolume(am.currentMusic.stream, am.volume)
	}
}
