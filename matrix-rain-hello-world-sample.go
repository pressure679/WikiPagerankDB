package main
import (
	"fmt"
	"math/rand"
	"sync"
	"os"
	"os/signal"
	"time"
	
	"github.com/gdamore/tcell"
)

// The TODO's: The relevant structs to make the mentioned (following TODO's) in are in the Stream struct.
// TODO: This shall have a []rune object to show a stream of sentence(s). - Streams are created to be shown in StreamDisplay. - The []rune object shall have an iterator as well. - l. 31
// TODO: Feed this condition with a true/false boolean value depended on whether or not a sentence is finished. - l. 73, l. 80
// - TODO: Give a stream (s) a new sentence when this condition is met. - l. 93
// - TODO: Here a new stream is set to be true, new stream is not started immediately, but when line 68 (if rand.Intn(100) < 66 {...}) feeds a new sentence. - l. 97
// TODO: Control the newColumn value here by a struct with a map with column as key and bool as item to define it is true. - l. l92

var curSizes sizes                   // current sizes
var curStreamsPerStreamDisplay = 0   // current amount of streams per display allowed
var sizesUpdateCh = make(chan sizes) // channel used to notify StreamDisplayManager
var streamDisplaysByColumn = make(map[int]*StreamDisplay)
var screen tcell.Screen

type sizes struct {
	width  int
	height int
}

// TODO: This shall have a []rune object to show a stream of sentence(s). - Streams a created to be shown in StreamDisplay. - The []rune object shall have an iterator as well.
type Stream struct {
	display      *StreamDisplay
	speed        int
	length       int
	headPos      int
	tailPos      int
	stopCh       chan bool
	headDone     bool
	sentence     []rune
	// sentenceDone bool
}
type StreamDisplay struct {
	column      int
	stopCh      chan bool
	streams     map[*Stream]bool
	streamsLock sync.Mutex
	newStream   chan bool
}

func (s *Stream) run() {
	blackStyle := tcell.StyleDefault

	// Make more colors when rune feeder for this function is defined
	midStyleA := blackStyle.Foreground(tcell.ColorGreen)
	midStyleB := blackStyle.Foreground(tcell.ColorLime)
	headStyleA := blackStyle.Foreground(tcell.ColorSilver)
	headStyleB := blackStyle.Foreground(tcell.ColorWhite)

	s.sentence = []rune{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'}

	var lastRune rune
	var counter int = 0
STREAM:
	for {
		select {
		case <-s.stopCh:
			break STREAM
		case <-time.After(time.Duration(s.speed) * time.Millisecond):
			if !s.headDone && s.headPos <= curSizes.height {
				newRune := s.sentence[counter]
				counter++

				// TODO: Feed this condition with a true/false boolean value depended on whether or not a sentence is finished
				if rand.Intn(100) < 66 {
					screen.SetCell(s.display.column, s.headPos-1, midStyleA, lastRune)
				} else {
					screen.SetCell(s.display.column, s.headPos-1, midStyleB, lastRune)
				}

				// TODO: Feed this condition with a true/false boolean value depended on whether or not a sentence is finished
				if rand.Intn(100) < 33 {
					screen.SetCell(s.display.column, s.headPos, headStyleA, newRune)
				} else {
					screen.SetCell(s.display.column, s.headPos, headStyleB, newRune)
				}
				lastRune = newRune
				s.headPos++
				if counter == len(s.sentence) { counter = 0 }
			} else {
				s.headDone = true
			}

			// TODO: Give a stream (s) a new sentence when this condition is met.
			// Comment: tailPos never decrements or is set to 0 (it only increments); the stream/object is deleted instead.
			if s.tailPos > 0 || s.headPos >= s.length {
				if s.tailPos == 0 {
					// TODO: Here a new stream is set to be true, new stream is not started immediately, but it is started when line 68 (if rand.Intn(100) < 66 {...}) feeds a new sentence.
					s.display.newStream <- true
				}
				if s.tailPos < curSizes.height {
					screen.SetCell(s.display.column, s.tailPos, blackStyle, ' ') //'\uFF60'
					s.tailPos++
				} else {
					break STREAM
				}
			}
		}
	}
	delete(s.display.streams, s)
}

// This function locks the Streams/StreamDisplay and starts new Streams if a newStream is received to this StreamDisplay (newStream indicates to start a new Stream in this StreamDisplay)
func (sd *StreamDisplay) run() {
	for {
		select {
		case <-sd.stopCh:
			sd.streamsLock.Lock()

			for s := range sd.streams {
				s.stopCh <- true
			}

			return

		case <-sd.newStream:
			time.Sleep(time.Duration(rand.Intn(9000)) * time.Millisecond)

			sd.streamsLock.Lock()

			s := &Stream{
				display: sd,
				stopCh:  make(chan bool),
				speed:   300 + rand.Intn(1100),
				length:  10 + rand.Intn(8), // length of a stream is between 10 and 18 runes
			}

			sd.streams[s] = true

			go s.run()

			sd.streamsLock.Unlock()
		}
	}
}

// Function to set new size to StreamDisplay if screen size changes (see main).
func setSizes(width int, height int) {
	s := sizes{
		width:  width,
		height: height,
	}
	curSizes = s
	curStreamsPerStreamDisplay = 1 + height/10
	sizesUpdateCh <- s
}

// This starts the StreamDisplay, checks if screen size changes; deletes/creates Streams if so, checks for key input; Stops/Clears program/StreamDisplay if so.
func main() {
	var err error

	rand.Seed(time.Now().UnixNano())

	if screen, err = tcell.NewScreen(); err != nil {
		fmt.Println("Cannot alloc screen, tcell.NewScreen() gave an error:\n%s", err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Println("Cannot start gomatrix, screen.Init() gave an error:\n%s", err)
		os.Exit(1)
	}
	screen.HideCursor()
	screen.Clear()

	go func() {
		var lastWidth int

		for newSizes := range sizesUpdateCh {
			diffWidth := newSizes.width - lastWidth

			if diffWidth == 0 { continue }

			if diffWidth > 0 {
				for newColumn := lastWidth; newColumn < newSizes.width; newColumn++ {
					// Here the new StreamDisplay is made and assigned to a column
					sd := &StreamDisplay{
						column:    newColumn,
						stopCh:    make(chan bool, 1),
						streams:   make(map[*Stream]bool),
						newStream: make(chan bool, 1), // will only be filled at start and when a spawning stream has it's tail released
					}
					// TODO: Control the newColumn value here by a struct with a map with column as key and bool as item to define it is true.
					streamDisplaysByColumn[newColumn] = sd

					go sd.run()

					sd.newStream <- true
				}
				lastWidth = newSizes.width
			}

			if diffWidth < 0 {
				for closeColumn := lastWidth - 1; closeColumn > newSizes.width; closeColumn-- {
					sd := streamDisplaysByColumn[closeColumn]

					delete(streamDisplaysByColumn, closeColumn)

					sd.stopCh <- true
				}
				lastWidth = newSizes.width
			}
		}
	}()

	setSizes(screen.Size())

	curFPS := 25
	fpsSleepTime := time.Duration(1000000/curFPS) * time.Microsecond
	go func() {
		for {
			time.Sleep(fpsSleepTime)
			screen.Show()
		}
	}()

	eventChan := make(chan tcell.Event)
	go func() {
		for {
			event := screen.PollEvent()
			eventChan <- event
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

EVENTS:
	for {
		select {
		case event := <-eventChan:
			switch ev := event.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyCtrlZ, tcell.KeyCtrlC:
					break EVENTS

				case tcell.KeyCtrlL:
					screen.Sync()

				case tcell.KeyRune:
					switch ev.Rune() {
					case 'q':
						break EVENTS

					case 'c':
						screen.Clear()
					}
				}
			case *tcell.EventResize: // set sizes
				w, h := ev.Size()
				setSizes(w, h)

			case *tcell.EventError: // quit
				break EVENTS
			}

		case signal := <-sigChan:
			fmt.Println("Have signal: \n%s", signal.String())
			break EVENTS
		}
	}

	screen.Fini()
}
