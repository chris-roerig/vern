package ui

import (
	"fmt"
	"os"
)

// Dracula theme ANSI colors
const (
	reset   = "\033[0m"
	green   = "\033[38;2;80;250;123m"  // #50fa7b - success
	yellow  = "\033[38;2;241;250;140m" // #f1fa8c - warning
	red     = "\033[38;2;255;85;85m"   // #ff5555 - error
	cyan    = "\033[38;2;139;233;253m" // #8be9fd - info
	pink    = "\033[38;2;255;121;198m" // #ff79c6 - accent
	purple  = "\033[38;2;189;147;249m" // #bd93f9 - highlight
	muted   = "\033[38;2;98;114;164m"  // #6272a4 - dim
)

func Success(format string, a ...any) {
	fmt.Fprintf(os.Stdout, green+format+reset+"\n", a...)
}

func Warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, yellow+format+reset+"\n", a...)
}

func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, red+format+reset+"\n", a...)
}

func Info(format string, a ...any) {
	fmt.Fprintf(os.Stdout, cyan+format+reset+"\n", a...)
}

func Accent(format string, a ...any) {
	fmt.Fprintf(os.Stdout, pink+format+reset+"\n", a...)
}

func Dim(format string, a ...any) {
	fmt.Fprintf(os.Stdout, muted+format+reset+"\n", a...)
}

const Logo = pink + `                    ▄▄         ▄▄   
▄▄ ▄▄ ▄▄▄▄ ▄▄▄▄     █   ▄▄  ▄▄  █   
██▄██ ██▄▄  ██▄█▄  ██   ███▄██  ██  
 ▀█▀  ██▄▄▄ ██ ██   █   ██ ▀██  █   
                    ▀▀         ▀▀   ` + reset
