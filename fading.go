package beepFade

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// Holds fadeItter and trackItter
/*
fadeItter - Is used to fade in and out
trackItter - Represents the position into a song
*/
var itters map[int][]float64

// fader is a type so that fader.Stream() can be used with proper parameters to run properly
type fader struct {
	// Streamer to fade
	Streamer beep.Streamer
	// How long in samples to fade in, and to fade out
	TimeSpan float64
	// What the volume should be for the streamer
	Volume float64
	// How long the audio is, so that fading in and out works properly
	audioLength float64
	// ID so that it can persist itterators between bits of slices
	id int
}

// For testing fading capabilities
func init() {
	// Necessary for itters map, otherwise there is a nil map error
	itters = make(map[int][]float64)
}

// Crossfades between all songs specified in files
func streamCreater(files ...string) {
	// Streamer that will contain all files
	var streamer beep.Streamer
	// Create 1000 samples of silence so that beep.Mix has a non-nil streamer to work with
	streamer = beep.Silence(1000)
	// The time span of the file previous to the one calculating on it. Used to get timing for crossfading right
	var lastTimeSpan float64
	// Specifies how long the streamer is, so that timing for crossfading is correct
	var position float64
	// Used so that speaker.Init has valid values for SampleRate, etc. Probably not a good idea if the SampleRates are different between files
	var format beep.Format
	// Iterate through all files specified to add them to streamer with proper crossfade
	for id, name := range files {
		// Open the file
		f, err := os.Open(name)
		if err != nil {
			fmt.Println("Couldn't find file " + name)
		}
		// Declared here so that format isn't specific to this block
		var s beep.StreamSeekCloser
		// Decode the file
		s, format, err = mp3.Decode(f)
		if err != nil {
			fmt.Println("Please ensure that " + name + " is an mp3, or is not corrupted")
		}
		// Create the set of parameters for it's stream function
		var faderStream = &fader{Streamer: s, Volume: 1, TimeSpan: float64(format.SampleRate.N(time.Second * 9)), audioLength: float64(s.Len()), id: id}
		// Create streamer with fading applied
		changedStreamer := beep.StreamerFunc(faderStream.Stream)
		// Create amount of silence before playing sound. Uses position, which by itself would make it play after the previous song. Subtracting lastTimeSpan makes a crossfade effect
		silenceAmount := int(position - lastTimeSpan)
		// Keeps previous streamer, and adds the new streamer with the silence in the beginning so it doesn't play over other songs
		streamer = beep.Mix(streamer, beep.Seq(beep.Silence(silenceAmount), changedStreamer))
		// Add position for next file
		position = position + faderStream.audioLength
		// Set last time span to current time span for next file
		lastTimeSpan = faderStream.TimeSpan
	}
	// Initialize speaker
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	// Create done channel so that program doesn't exit before all songs are played
	done := make(chan struct{})
	// Play streamer (doesn't belong here)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		close(done)
	})))
	// Wait until done is closed
	<-done
}

// Stream edits streamer so that it fades
func (v *fader) Stream(samples [][2]float64) (n int, ok bool) {
	// Determines if this specific streamer has been run before. If it hasn't then it needs to create fadeItter and trackItter for it
	if len(itters) < v.id+1 {
		// Print ID of song
		fmt.Println(v.id)
		// Create fadeItter and trackItter for the ID, and assign them to defaults of 0
		itters[v.id] = []float64{0, 0}
	}
	// Assign name to the map's ints for easier reading
	/*
		fadeItter - Is used to fade in and out
		trackItter - Represents the position into a song
	*/
	var fadeItter = &itters[v.id][0]
	var trackItter = &itters[v.id][1]
	// Use default streamer, and revise off of that
	n, ok = v.Streamer.Stream(samples)
	var gain float64
	gain = v.Volume
	// x1 is 0 and represents the start of the fade
	var x1 float64
	// The start of the fade should be silent, so y1 is 0
	var y1 float64
	// End point should be the TimeSpan set so that at the end of the TimeSpan, the gain is at requested value
	var x2 = v.TimeSpan
	// The requested gain, which will be played at the end of the TimeSpan
	var y2 = gain
	// Create the slope for a line representing this
	slopeUp := slopeCalc(x1, y1, x2, y2)
	//slopeDown := slopeCalc(x1, y2, x2, y1)
	// By default, sampleGain is the requested gain so between fadepoints, it is normal
	var sampleGain = gain
	// For each recieved sample, apply fade to it if necessary
	for i := range samples[:n] {
		// If the position in the track is after or at the time where it should begin to fade, then fade
		if *trackItter >= v.audioLength-v.TimeSpan {
			// Slope-intercept form to get gain
			/*
				m					x 							+ 	b
				Calculated slope	The position in the fade		The y intercept of the gain, so that it fades down from the gain
			*/
			sampleGain = -slopeUp*float64(*fadeItter) + gain
			// Increment fade so that the next iteration will reduce the gain by more
			*fadeItter++
			// Prevents possible bug where the gain may become negative, which will result in the song's gain becoming high again
			if sampleGain < 0 {
				sampleGain = 0
			}
			// If the position of the track is before the specified TimeSpan, and the fadeItter isn't above the TimeSpan, begin to fade in.
		} else if *trackItter <= v.TimeSpan && slopeUp*float64(*fadeItter) <= gain {
			// Slope-intercept form to get gain
			/*
				m					x 							+ 	b
				Calculated slope	The position in the fade		0, because it is fading in from nothing
			*/
			sampleGain = slopeUp * float64(*fadeItter)
			// Increment fade so that the next iteration will reduce the gain by more
			*fadeItter++
		} else {
			// Ensures fadeItter isn't already high from fading in when it is time to fade out
			*fadeItter = 0
		}
		// Set the samples to the calculated gain
		samples[i][0] *= sampleGain
		samples[i][1] *= sampleGain
		// Increment trackItter to update position in track
		*trackItter++
	}
	// Return the samples with gain applied, and whether or not operations were successful
	return n, ok
}

// Calculates the slope between two points
func slopeCalc(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	return (y2 - y1) / (x2 - x1)
}
