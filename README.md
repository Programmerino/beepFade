# beepFade
Adds a new fading effect for the Golang Beep library
### How to use
## Fading streamer
```go
// Create the set of parameters for it's stream function
// TimeSpan is how many samples to fade-in for, and then how many samples before the end to fade out for. It is set to 9 seconds here.
// audioLength is how long the audio it is playing is, which is necessary for fading out properly
// id should be a unique number so that it can stream while keeping some variable persistent.
var faderStream = &fader{Streamer: s, Volume: 1, TimeSpan: float64(format.SampleRate.N(time.Second * 9)), audioLength: float64(s.Len()), id: id}
// Create streamer with fading applied
changedStreamer := beep.StreamerFunc(faderStream.Stream)
```
## Crossfade between songs
```go
// You can use this high level function to crossfade between songs (must be mp3 right now)
streamCreater("song1.mp3", "song2.mp3")
// This will fade into song1, and then crossfade between song1 and song2, and then fade out of song2
```
