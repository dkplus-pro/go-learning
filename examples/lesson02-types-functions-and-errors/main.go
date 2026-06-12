package main

import "fmt"

func main() {
	input := SignupInput{
		Name:  " Ada Lovelace ",
		Email: "ada@example.com",
		Age:   28,
	}

	profile, err := BuildProfile(input)
	if err != nil {
		fmt.Println("build profile failed:", err)
		return
	}

	fmt.Printf("%s <%s> adult=%v\n", profile.DisplayName, profile.Email, profile.IsAdult)
}
