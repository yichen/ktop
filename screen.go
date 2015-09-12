package ktop

import (
	"log"

	"github.com/nsf/termbox-go"
)

type EventHandler func(keyEvent termbox.Event)

// Screen is the container of a full console screen
// it contains the data it news to draw as well
type Screen struct {
	// termbox InputMode
	InputMode termbox.InputMode

	// internal event channel for user input
	eventChan chan termbox.Event

	// context stack
	contexts []Context

	// exit channel
	ExitChan chan bool

	// control channel to stop the loop()
	stop chan struct{}
}

type Context interface {
	// key input event handlers
	OnKeyInput(screen Screen, keyEvent termbox.Event)

	// Refresh the screen content. This is where the content is printed on the console
	Refresh(screen Screen)

	// will Show(), called before the screen print the content
	WillShow(screen Screen)
}

func NewScreen(context Context) *Screen {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	return &Screen{
		InputMode: termbox.InputEsc | termbox.InputMouse,
		eventChan: make(chan termbox.Event, 16),
		ExitChan:  make(chan bool),
		stop:      make(chan struct{}),
		contexts:  []Context{context},
	}
}

func (s *Screen) refresh() {

	// clean up first before calling user refresh
	w, h := termbox.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	termbox.Flush()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	context := s.CurrentContext()
	context.Refresh(*s)
}

func (s *Screen) CurrentContext() Context {
	if len(s.contexts) == 0 {
		return nil
	}

	return s.contexts[len(s.contexts)-1]
}

func (s *Screen) Show() {
	context := s.CurrentContext()
	context.WillShow(*s)

	go s.loop()

	s.refresh()
	termbox.HideCursor()
	termbox.Flush()
}

func (s *Screen) WaitForExit() {
	<-s.ExitChan
	termbox.Close()
}

func (s *Screen) Print(text string, col int, row int, fg termbox.Attribute, bg termbox.Attribute) {
	for i, c := range text {
		termbox.SetCell(col+i, row, c, fg, bg)
	}
}

func (s *Screen) loop() {

	stopHandler := make(chan struct{})
	// start event handlers first to wait for events
	go s.handleEvents(stopHandler)

	// poll event from the global termbox event chan,and pass
	// it to the local channel
	go func() {
		for {
			select {
			case <-s.stop:
				stopHandler <- struct{}{}
				return

			default:
				ev := termbox.PollEvent()
				s.eventChan <- ev
			}
		}
	}()

}

func (s *Screen) handleEvents(stopHandler chan struct{}) {
	log.Println("Start handleEvents")

	for {
		context := s.CurrentContext()

		select {
		case ev := <-s.eventChan:
			if ev.Type == termbox.EventKey {
				context.OnKeyInput(*s, ev)
			}
			if ev.Type == termbox.EventError {
				panic(ev.Err)
			}
		case <-stopHandler:
			log.Println("Stopping handleEvents")
			return
		}
	}
}

func (s *Screen) Push(context Context) {
	// call interrupt so that the event loop is not waiting at termbox.PollEvent()
	// this way the s.stop signal can be picked up
	termbox.Interrupt()
	s.stop <- struct{}{}

	// push the current context to stack
	s.contexts = append(s.contexts, context)
	s.Show()
}

// Pop will close the current screen and goes back to the parent screen
func (s *Screen) Pop() {
	termbox.Interrupt()
	s.stop <- struct{}{}
	s.contexts = s.contexts[0 : len(s.contexts)-1]
	s.Show()
}
