package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/aktnb/pi-go-bot/commands"
	"github.com/aktnb/pi-go-bot/controller"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	s        *discordgo.Session
	c        *controller.Controller
	BotToken string
)

func init() {
	// .env ファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// 環境変数を取得
	BotToken = os.Getenv("DISCORD_BOT_TOKEN")

	// Discord セッションを作成
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	s.AddHandler(func(s *discordgo.Session, m *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v\n", m.User.Username, m.User.Discriminator)
	})
}

func main() {
	// コントローラーを作成
	c = controller.New()
	s.AddHandler(c.HandleInteraction)

	// Discord に接続
	err := s.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}
	defer s.Close()

	// コマンドを追加
	c.AddGlobalCommand(s, &commands.PingCommand)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Bot is now running. Press CTRL+C to exit.")
	<-stop
}
