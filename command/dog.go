package command

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aktnb/pi-go-bot/controller"

	"github.com/bwmarrin/discordgo"
)

type Dog struct {
	Url    string `json:"message"`
	Status string `json:"status"`
}

var DogCommand = controller.Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "dog",
		Description: "Get a random dog image",
	},
	Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		resp, err := http.Get("https://dog.ceo/api/breeds/image/random")

		if err != nil {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "U^q^U",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "U^q^U",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		body, _ := io.ReadAll(resp.Body)
		dog := Dog{}

		if err := json.Unmarshal(body, &dog); err != nil {
			log.Println(err)
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "U^q^U",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		data, err := http.DefaultClient.Get(dog.Url)
		if err != nil {
			log.Println(err)
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "U^q^U",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		defer data.Body.Close()

		if data.StatusCode != http.StatusOK {
			log.Println(err)
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "U^q^U",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Files: []*discordgo.File{
					{
						Name:        "dog.jpg",
						ContentType: "image/jpeg",
						Reader:      data.Body,
					},
				},
			},
		})
	},
}
