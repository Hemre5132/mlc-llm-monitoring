package database

import (
	"log"
	"time"

	"masterfabric-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// systemTopics, uygulama ilk açıldığında topics tablosuna eklenen 60 İngilizce
// essay konusudur. Kategorilere dengeli dağıtılmıştır (her kategoriden ~12 adet).
var systemTopics = []struct {
	Text       string
	Category   string
	Difficulty string
}{
	// ---- opinion (12) ----
	{Text: "Do you think social media has a positive or negative impact on society? Explain your opinion with examples.", Category: "opinion", Difficulty: "intermediate"},
	{Text: "Should students be required to wear school uniforms? Give reasons for your opinion.", Category: "opinion", Difficulty: "beginner"},
	{Text: "Is it better to live in a big city or a small town? Share your opinion and justify it.", Category: "opinion", Difficulty: "beginner"},
	{Text: "Do you believe that artificial intelligence will create more jobs than it eliminates? Explain.", Category: "opinion", Difficulty: "advanced"},
	{Text: "Should governments invest more in public transportation instead of building new highways? Argue your position.", Category: "opinion", Difficulty: "intermediate"},
	{Text: "Is remote work better for employees than working in an office? Give your opinion with reasons.", Category: "opinion", Difficulty: "intermediate"},
	{Text: "Do you think that money can buy happiness? Support your opinion with examples.", Category: "opinion", Difficulty: "beginner"},
	{Text: "Should junk food advertising be banned to protect children's health? Explain your view.", Category: "opinion", Difficulty: "intermediate"},
	{Text: "Is it important for young people to learn a second language? Share your opinion.", Category: "opinion", Difficulty: "beginner"},
	{Text: "Do you agree that university education should be free for everyone? Justify your answer.", Category: "opinion", Difficulty: "intermediate"},
	{Text: "Should companies be allowed to monitor their employees' online activity? Give your opinion.", Category: "opinion", Difficulty: "advanced"},
	{Text: "Is it better to be an early riser or a night owl? Explain your preference and reasoning.", Category: "opinion", Difficulty: "beginner"},

	// ---- narrative (12) ----
	{Text: "Describe a memorable journey you have taken and what made it special.", Category: "narrative", Difficulty: "beginner"},
	{Text: "Write about a time when you faced a difficult challenge and how you overcame it.", Category: "narrative", Difficulty: "intermediate"},
	{Text: "Tell the story of a day that changed your life forever.", Category: "narrative", Difficulty: "intermediate"},
	{Text: "Describe an unexpected act of kindness you witnessed or experienced.", Category: "narrative", Difficulty: "beginner"},
	{Text: "Write about a childhood memory that still makes you smile.", Category: "narrative", Difficulty: "beginner"},
	{Text: "Tell the story of how you met one of your closest friends.", Category: "narrative", Difficulty: "beginner"},
	{Text: "Describe a time when you had to make a difficult decision and what happened next.", Category: "narrative", Difficulty: "intermediate"},
	{Text: "Write about a place you visited that completely surprised you.", Category: "narrative", Difficulty: "intermediate"},
	{Text: "Tell the story of a mistake you made and the lesson you learned from it.", Category: "narrative", Difficulty: "intermediate"},
	{Text: "Describe a moment when you felt truly proud of yourself.", Category: "narrative", Difficulty: "beginner"},
	{Text: "Write about an experience that changed the way you see the world.", Category: "narrative", Difficulty: "advanced"},
	{Text: "Tell the story of a family tradition that is important to you.", Category: "narrative", Difficulty: "beginner"},

	// ---- argumentative (12) ----
	{Text: "Should zoos be banned? Present a strong argument for your position.", Category: "argumentative", Difficulty: "intermediate"},
	{Text: "Is social media doing more harm than good for teenagers? Argue your case.", Category: "argumentative", Difficulty: "intermediate"},
	{Text: "Should plastic bags be completely banned worldwide? Defend your argument.", Category: "argumentative", Difficulty: "intermediate"},
	{Text: "Do violent video games cause aggressive behavior in young people? Argue your position.", Category: "argumentative", Difficulty: "advanced"},
	{Text: "Should school start times be later for teenagers? Present your argument.", Category: "argumentative", Difficulty: "intermediate"},
	{Text: "Is it ethical to keep animals in captivity for entertainment? Argue your view.", Category: "argumentative", Difficulty: "advanced"},
	{Text: "Should voting be compulsory in democratic countries? Defend your position.", Category: "argumentative", Difficulty: "advanced"},
	{Text: "Are standardized tests a fair way to measure student ability? Argue your case.", Category: "argumentative", Difficulty: "intermediate"},
	{Text: "Should governments provide universal basic income to all citizens? Present your argument.", Category: "argumentative", Difficulty: "advanced"},
	{Text: "Is it better to protect the environment or prioritize economic growth? Argue your position.", Category: "argumentative", Difficulty: "intermediate"},
	{Text: "Should children be allowed to have smartphones? Defend your argument.", Category: "argumentative", Difficulty: "beginner"},
	{Text: "Is space exploration worth the cost? Present a convincing argument.", Category: "argumentative", Difficulty: "advanced"},

	// ---- descriptive (12) ----
	{Text: "Describe your favorite place to relax and explain why it is special to you.", Category: "descriptive", Difficulty: "beginner"},
	{Text: "Describe a person who has had a great influence on your life.", Category: "descriptive", Difficulty: "intermediate"},
	{Text: "Describe your dream house in detail, including the rooms and the atmosphere.", Category: "descriptive", Difficulty: "beginner"},
	{Text: "Describe a traditional dish from your country and how it is prepared.", Category: "descriptive", Difficulty: "intermediate"},
	{Text: "Describe a beautiful natural landscape you have seen, using vivid details.", Category: "descriptive", Difficulty: "intermediate"},
	{Text: "Describe your best friend and the qualities that make them special.", Category: "descriptive", Difficulty: "beginner"},
	{Text: "Describe a typical day in your life from morning to night.", Category: "descriptive", Difficulty: "beginner"},
	{Text: "Describe a historical building or monument you find fascinating.", Category: "descriptive", Difficulty: "intermediate"},
	{Text: "Describe the perfect holiday destination and what makes it ideal.", Category: "descriptive", Difficulty: "beginner"},
	{Text: "Describe a memorable meal you had and the atmosphere around it.", Category: "descriptive", Difficulty: "intermediate"},
	{Text: "Describe a city you know well, focusing on its sights, sounds, and atmosphere.", Category: "descriptive", Difficulty: "intermediate"},
	{Text: "Describe an object that holds great sentimental value for you.", Category: "descriptive", Difficulty: "beginner"},

	// ---- compare_contrast (12) ----
	{Text: "Compare and contrast living in the countryside with living in a big city.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast online learning with traditional classroom learning.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast reading a book with watching its movie adaptation.", Category: "compare_contrast", Difficulty: "beginner"},
	{Text: "Compare and contrast being an only child with growing up with siblings.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast public schools with private schools.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast summer and winter as your favorite seasons.", Category: "compare_contrast", Difficulty: "beginner"},
	{Text: "Compare and contrast working from home with working in an office.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast fast food with home-cooked meals.", Category: "compare_contrast", Difficulty: "beginner"},
	{Text: "Compare and contrast the experience of traveling alone with traveling in a group.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast traditional books with e-books.", Category: "compare_contrast", Difficulty: "beginner"},
	{Text: "Compare and contrast the advantages and disadvantages of social media.", Category: "compare_contrast", Difficulty: "intermediate"},
	{Text: "Compare and contrast life before and after the internet.", Category: "compare_contrast", Difficulty: "advanced"},
}

// SeedTopics, topics tablosu boşsa sistem konularını ekler.
// main.go'da database.Connect sonrası çağrılmalıdır.
func SeedTopics(db *gorm.DB) {
	var count int64
	if err := db.Model(&models.Topic{}).Where("source = ?", "system").Count(&count).Error; err != nil {
		log.Printf("SeedTopics: konu sayısı alınamadı: %v", err)
		return
	}

	if count > 0 {
		log.Printf("SeedTopics: sistem konuları zaten mevcut (%d adet), atlanıyor", count)
		return
	}

	topics := make([]models.Topic, 0, len(systemTopics))
	now := time.Now()
	for _, t := range systemTopics {
		topics = append(topics, models.Topic{
			ID:         uuid.New(),
			Text:       t.Text,
			Category:   t.Category,
			Difficulty: t.Difficulty,
			Source:     "system",
			CreatedAt:  now,
		})
	}

	if err := db.Create(&topics).Error; err != nil {
		log.Printf("SeedTopics: konular eklenemedi: %v", err)
		return
	}

	log.Printf("SeedTopics: %d sistem konusu eklendi", len(topics))
}

// GetDailyTopic, günün konusunu deterministik olarak seçer.
// Tüm kullanıcılar aynı gün aynı konuyu görür: index = (gün + yıl) % toplam_sistem_konu_sayısı
func GetDailyTopic(db *gorm.DB) (*models.Topic, error) {
	var topics []models.Topic
	if err := db.Where("source = ?", "system").Order("created_at asc").Find(&topics).Error; err != nil {
		return nil, err
	}

	if len(topics) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	now := time.Now()
	index := (now.Year() + now.YearDay()) % len(topics)
	return &topics[index], nil
}

// GetRandomTopic, random bir sistem konusu döner.
// excludeID verilirse o konu hariç tutulur (opsiyonel).
func GetRandomTopic(db *gorm.DB, excludeID *uuid.UUID) (*models.Topic, error) {
	query := db.Where("source = ?", "system")
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var topic models.Topic
	if err := query.Order("random()").First(&topic).Error; err != nil {
		return nil, err
	}

	return &topic, nil
}
