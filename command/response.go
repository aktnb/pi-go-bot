package command

import (
	"fmt"
	"strings"

	"github.com/aktnb/pi-go-bot/controller"
	"github.com/aktnb/pi-go-bot/service/response"

	"github.com/bwmarrin/discordgo"
)

func CustomResponseCommand(r *response.ResponseService) *controller.Command {
	return &controller.Command{
		ApplicationCommand: discordgo.ApplicationCommand{
			Name:        "custom-response",
			Description: "Set a custom response",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "Set a custom response",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "keyword",
							Description: "The keyword to trigger the response",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "response",
							Description: "The response to the keyword",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "delete",
					Description: "Delete a custom response",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "keyword",
							Description: "The keyword to delete the response",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List all custom responses",
				},
			},
		},
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			data := i.ApplicationCommandData()
			subCommand := data.Options[0].Name
			switch subCommand {
			case "set":
				keyword := data.Options[0].Options[0].StringValue()
				response := data.Options[0].Options[1].StringValue()
				err := r.SetResponse(i.Interaction.GuildID, keyword, response)
				if err != nil {
					return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Failed to set response",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				}
				return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("`%s` に対する応答を設定しました", keyword),
					},
				})
			case "delete":
				keyword := data.Options[0].Options[0].StringValue()
				err := r.DeleteResponse(i.Interaction.GuildID, keyword)
				if err != nil {
					return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Failed to delete response",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
				}
				return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("`%s` に対する応答を削除しました", keyword),
					},
				})
			case "list":
				var list []string
				c := 0
				for key := range r.GetKeys(i.Interaction.GuildID) {
					if c >= 25 {
						break
					}
					list = append(list, fmt.Sprintf("`%s`", key))
					c++
				}

				msg := "カスタムレスポンス一覧:\n" + strings.Join(list, " ")

				return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: msg,
					},
				})
			default:
				return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Unknown subcommand",
					},
				})
			}
		},
	}
}
