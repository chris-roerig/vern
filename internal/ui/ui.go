// Package ui provides Dracula-themed colored terminal output.
package ui

import (
	"fmt"
	"os"
)

const (
	reset  = "\033[0m"
	green  = "\033[38;2;80;250;123m"  // #50fa7b
	yellow = "\033[38;2;241;250;140m" // #f1fa8c
	red    = "\033[38;2;255;85;85m"   // #ff5555
	cyan   = "\033[38;2;139;233;253m" // #8be9fd
	pink   = "\033[38;2;255;121;198m" // #ff79c6
	muted  = "\033[38;2;98;114;164m"  // #6272a4
)

// Success prints a green success message to stdout.
func Success(format string, a ...any) {
	fmt.Fprintf(os.Stdout, green+format+reset+"\n", a...)
}

// Warn prints a yellow warning message to stderr.
func Warn(format string, a ...any) {
	fmt.Fprintf(os.Stderr, yellow+format+reset+"\n", a...)
}

// Error prints a red error message to stderr.
func Error(format string, a ...any) {
	fmt.Fprintf(os.Stderr, red+format+reset+"\n", a...)
}

// Info prints a cyan informational message to stdout.
func Info(format string, a ...any) {
	fmt.Fprintf(os.Stdout, cyan+format+reset+"\n", a...)
}

// Dim prints a muted/grey message to stdout.
func Dim(format string, a ...any) {
	fmt.Fprintf(os.Stdout, muted+format+reset+"\n", a...)
}

// Logo is the ver{n} ASCII art logo in pink.
const Logo = pink + `                    ▄▄         ▄▄   
▄▄ ▄▄ ▄▄▄▄ ▄▄▄▄     █   ▄▄  ▄▄  █   
██▄██ ██▄▄  ██▄█▄  ██   ███▄██  ██  
 ▀█▀  ██▄▄▄ ██ ██   █   ██ ▀██  █   
                    ▀▀         ▀▀   ` + reset
