package main

import "testing"

func TestGreet(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "with name", in: "Go", want: "Hello, Go!"},
		{name: "empty name", in: "", want: "Hello, friend!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Greet(tt.in)
			if got != tt.want {
				t.Fatalf("Greet(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
