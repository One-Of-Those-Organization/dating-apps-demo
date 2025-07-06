package main

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

func GetJWT(c *fiber.Ctx) (jwt.MapClaims, error) {
    user, ok := c.Locals("user").(*jwt.Token)
    if !ok || user == nil {
        return nil, errors.New("JWT token not valid")
    }
    if !user.Valid {
        return nil, errors.New("JWT token expired")
    }
    claims, ok := user.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("JWT claims not valid")
    }
    return claims, nil
}

