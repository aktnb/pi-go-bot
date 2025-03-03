package command

import (
	"github.com/aktnb/pi-go-bot/controller"

	"github.com/bwmarrin/discordgo"
)

var PingCommand = controller.Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Ping the bot to check if it's online",
	},
	Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
	},
}
