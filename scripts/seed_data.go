package main

import (
	"log"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	config.ConnectDB()

	// Seed visa data
	visas := []models.Visa{
		{
			Country:     "Japan",
			Type:        "Tourist",
			Description: "Tourist visa for Japan with 30 days validity",
			Price:       500000,
			Duration:    30,
			IsActive:    true,
		},
		{
			Country:     "South Korea",
			Type:        "Tourist",
			Description: "Tourist visa for South Korea with 30 days validity",
			Price:       400000,
			Duration:    30,
			IsActive:    true,
		},
		{
			Country:     "Singapore",
			Type:        "Tourist",
			Description: "Tourist visa for Singapore with 30 days validity",
			Price:       300000,
			Duration:    30,
			IsActive:    true,
		},
		{
			Country:     "Japan",
			Type:        "Business",
			Description: "Business visa for Japan with 90 days validity",
			Price:       800000,
			Duration:    90,
			IsActive:    true,
		},
	}

	for _, visa := range visas {
		if err := config.DB.Create(&visa).Error; err != nil {
			log.Printf("Failed to create visa %s: %v", visa.Country, err)
		} else {
			log.Printf("Created visa: %s - %s", visa.Country, visa.Type)
		}
	}

	// Seed visa options for Japan Tourist visa
	var japanTouristVisa models.Visa
	config.DB.Where("country = ? AND type = ?", "Japan", "Tourist").First(&japanTouristVisa)

	visaOptions := []models.VisaOption{
		{
			VisaID:      japanTouristVisa.ID,
			Name:        "Express Processing",
			Description: "Fast processing within 3-5 business days",
			Price:       200000,
			IsActive:    true,
		},
		{
			VisaID:      japanTouristVisa.ID,
			Name:        "Travel Insurance",
			Description: "Comprehensive travel insurance coverage",
			Price:       150000,
			IsActive:    true,
		},
		{
			VisaID:      japanTouristVisa.ID,
			Name:        "Airport Pickup",
			Description: "Airport pickup service in Japan",
			Price:       300000,
			IsActive:    true,
		},
	}

	for _, option := range visaOptions {
		if err := config.DB.Create(&option).Error; err != nil {
			log.Printf("Failed to create visa option %s: %v", option.Name, err)
		} else {
			log.Printf("Created visa option: %s", option.Name)
		}
	}

	log.Println("Seed data completed successfully!")
}
