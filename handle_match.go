package main

import (
	"dating-apps/table"
	"math"
	"math/rand"
	"sort"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Chromosome struct {
	UserA   table.User
	UserB   table.User
	Fitness float64
}

type Population []Chromosome

type MatchResult struct {
	User               table.User `json:"user"`
	CompatibilityScore float64   `json:"compatibility_score"`
}

func calculateFitness(userA, userB table.User) float64 {
	var score float64 = 0

	ageDiff := math.Abs(float64(userA.Age - userB.Age))
	ageScore := math.Max(0, 100-ageDiff*5)
	score += ageScore * 0.2

	// IDK THIS IS BAD
	if userA.Home == userB.Home {
		score += 30
	}

	sharedHobbies := countSharedItems(userA.Hobbies, userB.Hobbies)
	totalHobbies := len(userA.Hobbies) + len(userB.Hobbies)
	if totalHobbies > 0 {
		hobbyScore := float64(sharedHobbies*2) / float64(totalHobbies) * 100
		score += hobbyScore * 0.25
	}

	sharedInterests := countSharedInterests(userA.Interests, userB.Interests)
	totalInterests := len(userA.Interests) + len(userB.Interests)
	if totalInterests > 0 {
		interestScore := float64(sharedInterests*2) / float64(totalInterests) * 100
		score += interestScore * 0.25
	}

	if userA.Gender != userB.Gender {
		score += 10
	}

	return score
}

func countSharedItems(hobbiesA, hobbiesB []table.Hobby) int {
	shared := 0
	for _, hobbyA := range hobbiesA {
		for _, hobbyB := range hobbiesB {
			if hobbyA.ID == hobbyB.ID {
				shared++
				break
			}
		}
	}
	return shared
}

// Count shared interests between two users
func countSharedInterests(interestsA, interestsB []table.Interest) int {
	shared := 0
	for _, interestA := range interestsA {
		for _, interestB := range interestsB {
			if interestA.ID == interestB.ID {
				shared++
				break
			}
		}
	}
	return shared
}

// Create initial population
func createInitialPopulation(currentUser table.User, allUsers []table.User, populationSize int) Population {
	population := make(Population, 0, populationSize)

	for i := 0; i < populationSize && i < len(allUsers); i++ {
		if allUsers[i].ID != currentUser.ID {
			chromosome := Chromosome{
				UserA:   currentUser,
				UserB:   allUsers[i],
				Fitness: calculateFitness(currentUser, allUsers[i]),
			}
			population = append(population, chromosome)
		}
	}

	return population
}

func tournamentSelection(population Population, tournamentSize int) Chromosome {
	best := population[rand.Intn(len(population))]
	for i := 1; i < tournamentSize && i < len(population); i++ {
		competitor := population[rand.Intn(len(population))]
		if competitor.Fitness > best.Fitness {
			best = competitor
		}
	}
	return best
}

func crossover(parent1, parent2 Chromosome, allUsers []table.User) Chromosome {

	randomUser := allUsers[rand.Intn(len(allUsers))]

	child := Chromosome{
		UserA:   parent1.UserA,
		UserB:   randomUser,
		Fitness: calculateFitness(parent1.UserA, randomUser),
	}

	return child
}

func mutate(chromosome Chromosome, allUsers []table.User, mutationRate float64) Chromosome {
	if rand.Float64() < mutationRate {
		randomUser := allUsers[rand.Intn(len(allUsers))]
		if randomUser.ID != chromosome.UserA.ID {
			chromosome.UserB = randomUser
			chromosome.Fitness = calculateFitness(chromosome.UserA, randomUser)
		}
	}

	return chromosome
}

func runGeneticAlgorithm(currentUser table.User, allUsers []table.User, generations int) []MatchResult {
	populationSize := int(math.Min(50, float64(len(allUsers)-1)))
	if populationSize <= 0 {
		return []MatchResult{}
	}

	population := createInitialPopulation(currentUser, allUsers, populationSize)

	for generation := 0; generation < generations; generation++ {
		newPopulation := make(Population, 0, populationSize)

		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness > population[j].Fitness
		})

		eliteCount := populationSize / 5
		for i := 0; i < eliteCount && i < len(population); i++ {
			newPopulation = append(newPopulation, population[i])
		}

		for len(newPopulation) < populationSize {
			parent1 := tournamentSelection(population, 3)
			parent2 := tournamentSelection(population, 3)

			child := crossover(parent1, parent2, allUsers)
			child = mutate(child, allUsers, 0.1)

			newPopulation = append(newPopulation, child)
		}

		population = newPopulation
	}

	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	results := make([]MatchResult, 0, 5)
	seen := make(map[int]bool)

	for _, chromosome := range population {
		if len(results) >= 5 {
			break
		}

		if !seen[chromosome.UserB.ID] && chromosome.UserB.ID != currentUser.ID {
			seen[chromosome.UserB.ID] = true
			results = append(results, MatchResult{
				User:               chromosome.UserB,
				CompatibilityScore: chromosome.Fitness,
			})
		}
	}

	return results
}

// GET : api/p/matchmake
func HandleMatchmake(bend *Backend, route fiber.Router) {
	route.Get("matchmake", func(c *fiber.Ctx) error {
		claims, err := GetJWT(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code": fiber.StatusUnauthorized,
				"data": "Failed to claim JWT.",
			})
		}
		userName := claims["name"].(string)

		var currentUser table.User
		if err := bend.db.Preload("Hobbies").Preload("Interests").Where("user_name = ?", userName).First(&currentUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"code":    fiber.StatusNotFound,
					"message": "User not found",
					"data":    nil,
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"message": "Database error",
				"data":    nil,
			})
		}

		var allUsers []table.User
		if err := bend.db.Preload("Hobbies").Preload("Interests").Where("id != ?", currentUser.ID).Find(&allUsers).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"message": "Failed to fetch users",
				"data":    nil,
			})
		}

		if len(allUsers) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": fiber.StatusBadRequest,
				"data": "No other users available for matching.",
			})
		}

		matches := runGeneticAlgorithm(currentUser, allUsers, 50)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "Matchmaking completed successfully",
			"data":    matches,
		})
	})
}
