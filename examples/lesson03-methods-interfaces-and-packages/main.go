package main

import (
	"fmt"
	"log"

	"github.com/dkplus-pro/go-learning/examples/lesson03-methods-interfaces-and-packages/internal/learning"
)

func main() {
	store := learning.NewMemoryStore()

	learner := learning.Learner{Name: "Frontend Architect", CompletedLessons: 2}
	learner.CompleteLesson()

	if err := store.Save(learner); err != nil {
		log.Fatal(err)
	}

	saved, err := store.FindByName("Frontend Architect")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s completed %d lessons and is %s\n", saved.Name, saved.CompletedLessons, saved.Level())
}
