package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    // The cost parameter determines how computationally expensive the hash is to calculate
    // The default is 10, but you can increase it for better security (at the cost of performance)
    hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hashedBytes), nil
}

func CheckPassword(hashedPassword, plainPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
    return err == nil
}

