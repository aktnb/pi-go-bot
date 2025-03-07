package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/aktnb/pi-go-bot/command"
	"github.com/aktnb/pi-go-bot/controller"
	"github.com/aktnb/pi-go-bot/service/room"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	s        *discordgo.Session
	c        *controller.Controller
	db       *sql.DB
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
}

func init() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	var err error
	// PostgreSQL に接続
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalln(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalln(err)
	}
	log.Println("Connected to the database")
}

func init() {
	var err error
	// Discord セッションを作成
	s, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	// コントローラーを作成
	c = controller.New()
	s.AddHandler(c.HandleInteraction)

	r := room.New(db)

	s.Identify.Intents |= discordgo.IntentsGuildPresences

	// イベントハンドラーを追加
	s.AddHandler(func(s *discordgo.Session, m *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v\n", m.User.Username, m.User.Discriminator)
	})
	s.AddHandler(r.HandleGuildCreate)
	s.AddHandler(r.HandleVoiceStateUpdate)

	// Discord に接続
	defer s.Close()
	err := s.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}

	// コマンドを追加
	c.AddGlobalCommand(s, &command.PingCommand)
	c.AddGlobalCommand(s, &command.CatCommand)
	c.AddGlobalCommand(s, &command.DogCommand)

	// スラッシュコマンドを登録
	if err := c.OverwriteGlobalCommands(s); err != nil {
		log.Fatalf("Error overwriting global commands: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Bot is now running. Press CTRL+C to exit.")
	<-stop
}
