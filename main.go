package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

var s *discordgo.Session

var (
	BotToken string
)

func init() {
	BotToken = os.Getenv("DISCORD_BOT_TOKEN")
}

func main() {
	var err error
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, m *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v\n", m.User.Username, m.User.Discriminator)
	})

	err = s.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Bot is now running. Press CTRL+C to exit.")
	<-stop
}
