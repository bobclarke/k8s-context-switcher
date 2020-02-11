package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"unicode"

	"github.com/nsf/termbox-go"
)

var debugMode bool = false
var inputCharCount int = 0
var fg = termbox.ColorDefault
var bg = termbox.ColorDefault

const inputWidth int = 40
const inputHeight int = 2
const inputX int = 5
const inputY int = 5

type Contexts struct {
	selected_context       int
	context_array_all      []string
	context_array_filtered []string
	searchString           string
	currentContext         string
	SetOutput              string
}

func main() {
	contexts := &Contexts{}
	contexts.selected_context = 0
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputAlt | termbox.InputMouse)
	drawInputBox(fg, bg)
	contexts.getContexts()
	contexts.displayCurrentContext()
	contexts.displayOutputText(fg, bg)
	termbox.Flush()
	mainLoop(contexts)
}

func mainLoop(contexts *Contexts) {
mainloop:
	for {
		e := termbox.PollEvent()

		if e.Key == termbox.KeyArrowDown {
			contexts.selected_context++
		}
		if e.Key == termbox.KeyArrowUp {
			contexts.selected_context--
		}

		// Append the last keypress to the search string
		if !unicode.IsControl(e.Ch) {
			contexts.searchString += string(e.Ch)
		}

		// Handle backspace
		if e.Key == 127 {
			if len(contexts.searchString) > 0 {
				contexts.searchString = contexts.searchString[:len(contexts.searchString)-1]
			}
		}

		contexts.displayInputText(&e, fg, bg)
		contexts.filterContexts(&e)
		contexts.clearOutputText()
		contexts.displayOutputText(fg, bg)

		if e.Key == 13 {
			contexts.useContext()
			contexts.displayCurrentContext()
		}

		debug(e, contexts)

		//if e.Key == 13 || e.Key == 3 {
		if e.Key == 3 {
			break mainloop
		}

		termbox.Flush()
	}
}

func (c *Contexts) useContext() {
	ctx := c.context_array_filtered[c.selected_context]
	out1, err1 := exec.Command("kubectl", "config", "use-context", ctx).Output()
	if err1 != nil {
		log.Fatal(err1)
	}
	c.SetOutput = string(out1)

}

func (c *Contexts) displayCurrentContext() {

	out2, err2 := exec.Command("kubectl", "config", "current-context").Output()
	if err2 != nil {
		log.Fatal(err2)
	}
	c.currentContext = string(out2)

	x := inputX
	y := inputY - 2

	output := fmt.Sprintf("Current ctx: %s", c.currentContext)

	for i := 0; i <= 100; i++ {
		termbox.SetCell(i, y, ' ', fg, bg)
	}

	for _, c := range output {
		termbox.SetCell(x, y, c, termbox.ColorMagenta, bg)
		x++
	}
}

func (c *Contexts) filterContexts(e *termbox.Event) {
	c.context_array_filtered = nil
	for _, context := range c.context_array_all {
		if strings.Contains(context, c.searchString) {
			c.context_array_filtered = append(c.context_array_filtered, context)
		}
	}
}

func (c *Contexts) displayInputText(e *termbox.Event, fg termbox.Attribute, bg termbox.Attribute) {
	s := c.searchString

	for i := inputX - 5; i < inputWidth; i++ {
		termbox.SetCell(i+inputX, inputY+1, ' ', fg, bg) // Clear input box
	}

	for i, ch := range s {
		termbox.SetCell(i+inputX, inputY+1, ch, fg, bg) // Write search string to inout box
	}
}

func (c Contexts) clearOutputText() {
	for y := inputY + 5; y < len(c.context_array_all); y++ {

		for x := inputX; x < 100; x++ {
			termbox.SetCell(x, y, ' ', fg, bg)
		}
	}
}

func (c *Contexts) displayOutputText(fg termbox.Attribute, bg termbox.Attribute) {
	y := inputY + 5
	savedBg := bg

	for i, context := range c.context_array_filtered {
		if i == c.selected_context {
			bg = termbox.ColorYellow
		}
		x := inputX
		for _, c := range context {
			termbox.SetCell(x, y, c, fg, bg)
			x++
		}
		bg = savedBg
		y++
	}
}

func drawInputBox(fg termbox.Attribute, bg termbox.Attribute) {
	for i := inputX; i < (inputWidth + inputX); i++ {
		termbox.SetCell(i, inputY, rune('-'), fg, bg)
	}

	for i := inputX; i < (inputWidth + inputX); i++ {
		termbox.SetCell(i, inputY+2, rune('-'), fg, bg)
	}

	termbox.SetCell(inputX-1, inputY+1, rune('|'), fg, bg)
	termbox.SetCell(inputX+inputWidth, inputY+1, rune('|'), fg, bg)
	termbox.SetCell(inputX-1, inputY, rune('┌'), fg, bg)
	termbox.SetCell(inputX-1, inputY+2, rune('└'), fg, bg)
	termbox.SetCell(inputX+inputWidth, inputY, rune('┐'), fg, bg)
	termbox.SetCell(inputX+inputWidth, inputY+2, rune('┘'), fg, bg)
}

func (c *Contexts) getContexts() {
	out, err := exec.Command("kubectl", "config", "get-contexts", "-o", "name").Output()
	if err != nil {
		log.Fatal(err)
	}

	c.context_array_all = (strings.Split(string(out), "\n"))
}

func debug(e termbox.Event, c *Contexts) {
	if !debugMode {
		return
	}

	for i := inputX; i < 200; i++ {
		termbox.SetCell(i, inputY-2, ' ', fg, bg) // Clear debug display
	}

	//output := fmt.Sprintf("EventKey: %d Character: %c Search String %s Search String length %d",
	//	e.Key, e.Ch, c.searchString, len(c.searchString))

	output := fmt.Sprintf("Selected: %s Current: %s Set Output %s",
		c.context_array_filtered[c.selected_context],
		c.currentContext,
		c.SetOutput)

	x := inputX
	for _, c := range output {
		termbox.SetCell(x, inputY-2, c, fg, bg)
		x++
	}
}
