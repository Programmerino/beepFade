# beepFade
Adds a new fading effect for the Golang Beep library
## How to use
### Crossfade between songs
```go
// CrossfadeStream crossfades between all songs specified in files
// The sample-rates between the two streams must be the same, otherwise weird things might happen
// If opts is nil, then reasonable defaults are used
streamer := beepFade.CrossfadeStream(format, nil, stream1, stream2)
// This streamer will fade into stream1, and then crossfade between stream1 and stream2, and then fade out of stream2
```
A full example can be found in the example folder
