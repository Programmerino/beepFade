package main

import (
	"log"
	"os"
	"time"

	"github.com/Programmerino/beepFade"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	f, err := os.Open("song1.mp3")
	if err != nil {
		log.Fatal(err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	f2, err := os.Open("song2.mp3")
	if err != nil {
		log.Fatal(err)
	}
	streamer2, _, err := mp3.Decode(f2)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer2.Close()

	opts := beepFade.Options{
		TimeSpan: 20 * time.Second, // this setting sounds terrible, but hey, this is just an example
		Volume:   1,
	}

	final := beepFade.CrossfadeStream(format, &opts, streamer, streamer2)

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(final, beep.Callback(func() {
		done <- true
	})))

	<-done

}
