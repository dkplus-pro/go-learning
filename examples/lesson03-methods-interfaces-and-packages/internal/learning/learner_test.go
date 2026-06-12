package learning

import "testing"

func TestLearnerLevel(t *testing.T) {
	tests := []struct {
		name      string
		completed int
		want      string
	}{
		{name: "beginner", completed: 0, want: "beginner"},
		{name: "intermediate", completed: 3, want: "intermediate"},
		{name: "advanced", completed: 8, want: "advanced"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			learner := Learner{Name: "Ada", CompletedLessons: tt.completed}
			if got := learner.Level(); got != tt.want {
				t.Fatalf("Level() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompleteLesson(t *testing.T) {
	learner := Learner{Name: "Ada", CompletedLessons: 1}

	learner.CompleteLesson()

	if learner.CompletedLessons != 2 {
		t.Fatalf("CompletedLessons = %d, want 2", learner.CompletedLessons)
	}
}
