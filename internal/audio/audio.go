package audio

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AudioManager struct {
	MenuMusic    Music
	GameMusic    Music
	GameOverSFX  Sound
	CollectSFX   Sound
	Volume       float32
	CurrentMusic *Music
	IsPlaying    bool // Add playing status
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
		Volume: 1.0,
	}
}

func (am *AudioManager) LoadResources() {
	// Load menu music
	menuStream := rl.LoadMusicStream("assets/mainmenu.mp3")
	if !rl.IsMusicValid(menuStream) {
		fmt.Println("Failed to load menu music")
		am.MenuMusic = Music{stream: menuStream, loaded: false}
	} else {
		fmt.Println("Menu music loaded successfully")
		am.MenuMusic = Music{stream: menuStream, loaded: true}

	}

	// Load game music
	gameStream := rl.LoadMusicStream("assets/gamemusic.mp3")
	if !rl.IsMusicValid(gameStream) {
		fmt.Println("Failed to load game music")
		am.GameMusic = Music{stream: gameStream, loaded: false}
	} else {
		fmt.Println("Game music loaded successfully")
		am.GameMusic = Music{stream: gameStream, loaded: true}
	}

	// Load sound effects
	gameOverSound := rl.LoadSound("assets/gameover.wav")
	am.GameOverSFX = Sound{sound: gameOverSound, loaded: true}

	collectSound := rl.LoadSound("assets/nom.wav")
	rl.SetSoundVolume(collectSound, am.Volume*0.5)
	am.CollectSFX = Sound{sound: collectSound, loaded: true}

	// Set initial properties
	rl.SetMusicVolume(gameStream, am.Volume)
	rl.SetMusicPitch(gameStream, 1.0)
}

func (am *AudioManager) UnloadResources() {
	// Unload music
	if am.MenuMusic.loaded {
		rl.UnloadMusicStream(am.MenuMusic.stream)
	}
	if am.GameMusic.loaded {
		rl.UnloadMusicStream(am.GameMusic.stream)
	}

	// Unload sound effects
	if am.GameOverSFX.loaded {
		rl.UnloadSound(am.GameOverSFX.sound)
	}
	if am.CollectSFX.loaded {
		rl.UnloadSound(am.CollectSFX.sound)
	}

	rl.CloseAudioDevice()
}

func (am *AudioManager) PlayMusic(music *Music) {
	if music == nil || !music.loaded {
		fmt.Println("Attempted to play invalid music")
		return
	}

	// Stop current music if playing
	if am.CurrentMusic != nil && am.CurrentMusic.loaded {
		fmt.Println("Stopping current music")
		rl.StopMusicStream(am.CurrentMusic.stream)
		am.IsPlaying = false
	}

	am.CurrentMusic = music
	fmt.Printf("Playing new music (loaded: %v)\n", music.loaded)

	if rl.IsMusicValid(music.stream) {
		rl.SeekMusicStream(music.stream, 0.0)
		rl.PlayMusicStream(music.stream)
		rl.SetMusicVolume(music.stream, am.Volume)
		am.IsPlaying = true
		fmt.Println("Music started successfully")
	} else {
		fmt.Println("Failed to play music - stream not ready")
	}
}

func (am *AudioManager) UpdateMusic() {
	if am.CurrentMusic == nil || !am.CurrentMusic.loaded {
		return
	}

	if !rl.IsMusicStreamPlaying(am.CurrentMusic.stream) && am.IsPlaying {
		fmt.Println("Music ended, restarting...")
		rl.SeekMusicStream(am.CurrentMusic.stream, 0.0)
		rl.PlayMusicStream(am.CurrentMusic.stream)
	}

	rl.UpdateMusicStream(am.CurrentMusic.stream)
}

func (am *AudioManager) PlaySound(sound *Sound) {
	if sound.loaded {
		rl.PlaySound(sound.sound)
	}
}

func (am *AudioManager) SetVolume(volume float32) {
	am.Volume = volume / 100.0
	rl.SetMasterVolume(am.Volume)
	// Also update current music volume if playing
	if am.CurrentMusic != nil && am.CurrentMusic.loaded {
		rl.SetMusicVolume(am.CurrentMusic.stream, am.Volume)
	}
}
