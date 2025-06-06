package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    hash := "$2a$10$1CBH45rwl3OXdNHH9Sgt9O.MmHHvxjpd5uZIm9FT5lccRHB1HlRQC"
    password := "Test1234"

    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        fmt.Println("❌ Password mismatch:", err)
    } else {
        fmt.Println("✅ Password OK")
    }
}

