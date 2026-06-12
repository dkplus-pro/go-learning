package main

import (
	"errors"
	"testing"
)

func TestBuildProfile(t *testing.T) {
	tests := []struct {
		name    string
		input   SignupInput
		want    UserProfile
		wantErr bool
	}{
		{
			name:  "valid input trims name and normalizes email",
			input: SignupInput{Name: " Ada ", Email: "ADA@EXAMPLE.COM ", Age: 28},
			want:  UserProfile{DisplayName: "Ada", Email: "ada@example.com", IsAdult: true},
		},
		{
			name:  "teen user is valid but not adult",
			input: SignupInput{Name: "Grace", Email: "grace@example.com", Age: 16},
			want:  UserProfile{DisplayName: "Grace", Email: "grace@example.com", IsAdult: false},
		},
		{
			name:    "name is required",
			input:   SignupInput{Name: " ", Email: "ada@example.com", Age: 28},
			wantErr: true,
		},
		{
			name:    "email must contain at sign",
			input:   SignupInput{Name: "Ada", Email: "ada.example.com", Age: 28},
			wantErr: true,
		},
		{
			name:    "age must be at least thirteen",
			input:   SignupInput{Name: "Ada", Email: "ada@example.com", Age: 12},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildProfile(tt.input)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("BuildProfile() error = %v, want ErrInvalidInput", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("BuildProfile() unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("BuildProfile() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
