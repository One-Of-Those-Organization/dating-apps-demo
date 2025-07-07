package main

import (
	"dating-apps/table"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

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

// Nominatim API response structure
type NominatimResponse struct {
	PlaceID     int      `json:"place_id"`
	DisplayName string   `json:"display_name"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
}

// LocationData holds geocoded location information
type LocationData struct {
	Lat         float64
	Lon         float64
}

// Cache for geocoded locations to avoid repeated API calls
var locationCache = make(map[string]LocationData)

// Geocode location using Nominatim API
func geocodeLocation(location string) (LocationData, error) {
	// Check cache first
	if cached, exists := locationCache[strings.ToLower(location)]; exists {
		return cached, nil
	}

	// Prepare URL
	baseURL := "https://nominatim.openstreetmap.org/search"
	params := url.Values{}
	params.Add("q", location)
	params.Add("format", "json")
	params.Add("polygon", "1")
	params.Add("addressdetails", "1")
	params.Add("limit", "1")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make request
	resp, err := client.Get(fullURL)
	if err != nil {
		return LocationData{}, err
	}
	defer resp.Body.Close()

	// Parse response
	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return LocationData{}, err
	}

	if len(results) == 0 {
		return LocationData{}, fmt.Errorf("no results found for location: %s", location)
	}

	result := results[0]

	// Parse coordinates
	lat, err := strconv.ParseFloat(result.Lat, 64)
	if err != nil {
		return LocationData{}, err
	}

	lon, err := strconv.ParseFloat(result.Lon, 64)
	if err != nil {
		return LocationData{}, err
	}

	locationData := LocationData{
		Lat:         lat,
		Lon:         lon,
	}

	// Cache the result
	locationCache[strings.ToLower(location)] = locationData

	return locationData, nil
}

// Calculate distance between two coordinates using Haversine formula
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

// Calculate location compatibility score based on distance
func calculateLocationScore(userA, userB table.User) float64 {
	// If locations are exactly the same, give full score
	if strings.EqualFold(userA.Home, userB.Home) {
		return 30.0
	}

	// Try to geocode both locations
	locationA, errA := geocodeLocation(userA.Home)
	locationB, errB := geocodeLocation(userB.Home)

	// If geocoding fails, fall back to string comparison
	if errA != nil || errB != nil {
		if userA.Home == userB.Home {
			return 30.0
		}
		return 0.0
	}

	// Calculate distance between locations
	distance := calculateDistance(locationA.Lat, locationA.Lon, locationB.Lat, locationB.Lon)

	// Convert distance to score (closer = higher score)
	// Score decreases exponentially with distance
	// Max score: 30, decreases to near 0 at 100km+
	score := 30.0 * math.Exp(-distance/50.0)

	return score
}

func calculateFitness(userA, userB table.User) float64 {
	var score float64 = 0

	// Age compatibility (20% weight)
	ageDiff := math.Abs(float64(userA.Age - userB.Age))
	ageScore := math.Max(0, 100-ageDiff*5)
	score += ageScore * 0.2

	// Location compatibility using distance-based scoring
	locationScore := calculateLocationScore(userA, userB)
	score += locationScore

	// Shared hobbies (25% weight)
	sharedHobbies := countSharedItems(userA.Hobbies, userB.Hobbies)
	totalHobbies := len(userA.Hobbies) + len(userB.Hobbies)
	if totalHobbies > 0 {
		hobbyScore := float64(sharedHobbies*2) / float64(totalHobbies) * 100
		score += hobbyScore * 0.25
	}

	// Shared interests (25% weight)
	sharedInterests := countSharedInterests(userA.Interests, userB.Interests)
	totalInterests := len(userA.Interests) + len(userB.Interests)
	if totalInterests > 0 {
		interestScore := float64(sharedInterests*2) / float64(totalInterests) * 100
		score += interestScore * 0.25
	}

	// Gender diversity bonus
	if userA.Gender != userB.Gender {
		score += 10
	} else { score -= 1e12 }

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

// Create initial population - fixed for genetic algorithm with small datasets
func createInitialPopulation(currentUser table.User, allUsers []table.User, populationSize int) Population {
	population := make(Population, 0, populationSize)

	// Get all other users (excluding current user)
	otherUsers := make([]table.User, 0)
	for _, user := range allUsers {
		if user.ID != currentUser.ID {
			otherUsers = append(otherUsers, user)
		}
	}

	if len(otherUsers) == 0 {
		return population
	}

	// Ensure minimum population size for genetic algorithm
	minPopSize := int(math.Max(10, float64(len(otherUsers)*5)))
	if populationSize < minPopSize {
		populationSize = minPopSize
	}

	// Create chromosomes by repeating other users to fill population
	for i := 0; i < populationSize; i++ {
		selectedUser := otherUsers[i%len(otherUsers)]
		chromosome := Chromosome{
			UserA:   currentUser,
			UserB:   selectedUser,
			Fitness: calculateFitness(currentUser, selectedUser),
		}
		population = append(population, chromosome)
	}

	// Add some randomness to the initial population
	for i := 0; i < len(population)/2; i++ {
		randomUser := otherUsers[rand.Intn(len(otherUsers))]
		population[i].UserB = randomUser
		population[i].Fitness = calculateFitness(currentUser, randomUser)
	}

	return population
}

func tournamentSelection(population Population, tournamentSize int) Chromosome {
	if len(population) == 0 {
		return Chromosome{}
	}

	// Adjust tournament size for small populations
	actualTournamentSize := max(int(math.Min(float64(tournamentSize), float64(len(population)))), 1)

	best := population[rand.Intn(len(population))]
	for i := 1; i < actualTournamentSize; i++ {
		competitor := population[rand.Intn(len(population))]
		if competitor.Fitness > best.Fitness {
			best = competitor
		}
	}
	return best
}

func crossover(parent1, parent2 Chromosome, allUsers []table.User) Chromosome {
	// Get other users (excluding current user)
	otherUsers := make([]table.User, 0)
	for _, user := range allUsers {
		if user.ID != parent1.UserA.ID {
			otherUsers = append(otherUsers, user)
		}
	}

	if len(otherUsers) == 0 {
		return parent1
	}

	// Choose between parent1.UserB and parent2.UserB, or pick random
	var selectedUser table.User
	choice := rand.Intn(3)

	switch choice {
	case 0:
		selectedUser = parent1.UserB
	case 1:
		selectedUser = parent2.UserB
	case 2:
		selectedUser = otherUsers[rand.Intn(len(otherUsers))]
	}

	child := Chromosome{
		UserA:   parent1.UserA,
		UserB:   selectedUser,
		Fitness: calculateFitness(parent1.UserA, selectedUser),
	}

	return child
}

func mutate(chromosome Chromosome, allUsers []table.User, mutationRate float64) Chromosome {
	if rand.Float64() < mutationRate {
		// Get other users (excluding current user)
		otherUsers := make([]table.User, 0)
		for _, user := range allUsers {
			if user.ID != chromosome.UserA.ID {
				otherUsers = append(otherUsers, user)
			}
		}

		if len(otherUsers) > 0 {
			randomUser := otherUsers[rand.Intn(len(otherUsers))]
			chromosome.UserB = randomUser
			chromosome.Fitness = calculateFitness(chromosome.UserA, randomUser)
		}
	}

	return chromosome
}

func runGeneticAlgorithm(currentUser table.User, allUsers []table.User, generations int) []MatchResult {
	// Get other users first
	otherUsers := make([]table.User, 0)
	for _, user := range allUsers {
		if user.ID != currentUser.ID {
			otherUsers = append(otherUsers, user)
		}
	}

	if len(otherUsers) == 0 {
		return []MatchResult{}
	}

	// Adjust population size based on available users
	// For small datasets, create larger population with repetitions
	basePopSize := 50
	if len(otherUsers) < 10 {
		basePopSize = len(otherUsers) * 10 // Create 10 chromosomes per other user
	}
	populationSize := int(math.Max(float64(basePopSize), float64(len(otherUsers)*5)))

	population := createInitialPopulation(currentUser, allUsers, populationSize)

	if len(population) == 0 {
		return []MatchResult{}
	}

	// Run genetic algorithm
	for range generations {
		newPopulation := make(Population, 0, populationSize)

		// Sort by fitness (descending)
		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness > population[j].Fitness
		})

		// Elitism - keep top 20% of population
		eliteCount := int(math.Max(1, float64(populationSize)*0.2))
		for i := 0; i < eliteCount && i < len(population); i++ {
			newPopulation = append(newPopulation, population[i])
		}

		// Generate offspring to fill remaining population
		for len(newPopulation) < populationSize {
			// Tournament selection
			tournamentSize := int(math.Max(2, float64(len(population))*0.1))
			parent1 := tournamentSelection(population, tournamentSize)
			parent2 := tournamentSelection(population, tournamentSize)

			// Crossover
			child := crossover(parent1, parent2, allUsers)

			// Mutation
			child = mutate(child, allUsers, 0.15) // Slightly higher mutation rate for small populations

			newPopulation = append(newPopulation, child)
		}

		population = newPopulation
	}

	// Sort final population by fitness
	sort.Slice(population, func(i, j int) bool {
		return population[i].Fitness > population[j].Fitness
	})

	// Extract top unique matches
	results := make([]MatchResult, 0, 5)
	seen := make(map[int]bool)

	for _, chromosome := range population {
		if len(results) >= 5 {
			break
		}

		// Ensure we don't include the current user and avoid duplicates
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
				"data":    err.Error(),
			})
		}

		var allUsers []table.User
		if err := bend.db.Preload("Hobbies").Preload("Interests").Find(&allUsers).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    fiber.StatusInternalServerError,
				"message": "Failed to fetch users",
				"data":    err.Error(),
			})
		}

		// Count other users
		otherUsersCount := 0
		for _, user := range allUsers {
			if user.ID != currentUser.ID {
				otherUsersCount++
			}
		}

		if otherUsersCount == 0 {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"code":    fiber.StatusOK,
				"message": "No other users available for matching",
				"data":    []MatchResult{},
			})
		}

		// Run genetic algorithm
		matches := runGeneticAlgorithm(currentUser, allUsers, 50)

		var deletionList []int
		for i, m := range matches {
			if m.CompatibilityScore <= 0 {
				deletionList = append(deletionList, i)
			}
		}
		slices.Reverse(deletionList)
		for _, deleteIndex := range deletionList {
			matches = append(matches[:deleteIndex], matches[deleteIndex+1:]...)
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "Matchmaking completed successfully",
			"data":    matches,
		})
	})
}
