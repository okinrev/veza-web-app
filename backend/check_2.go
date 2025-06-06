package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$gPFVka9twToCAb3CMGaCPOXLHjzzMqaocun77d2iHP4o0N/zkqbTS"
	password := "Azerty1232"

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Println("❌ Mismatch:", err)
	} else {
		fmt.Println("✅ Password OK")
	}
}

