package command

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/aktnb/pi-go-bot/controller"

	"github.com/bwmarrin/discordgo"
)

type Cat struct {
	Id     string `json:"id"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

var CatCommand = controller.Command{
	ApplicationCommand: discordgo.ApplicationCommand{
		Name:        "cat",
		Description: "Get a random cat image",
	},
	Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		var resp *http.Response
		var err error
		errRespond := func() error {
			log.Printf("Failed to get cat image %v", err)
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "(=^・^=)",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		if resp, err = http.Get("https://api.thecatapi.com/v1/images/search"); err != nil {
			return errRespond()
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errRespond()
		}

		body, _ := io.ReadAll(resp.Body)
		cats := []Cat{}

		if err := json.Unmarshal(body, &cats); err != nil {
			return errRespond()
		}

		data, err := http.DefaultClient.Get(cats[0].Url)
		if err != nil {
			return errRespond()
		}
		defer data.Body.Close()

		if data.StatusCode != http.StatusOK {
			return errRespond()
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Files: []*discordgo.File{
					{
						Name:        "cat.jpg",
						ContentType: "image/jpeg",
						Reader:      data.Body,
					},
				},
			},
		})
	},
}
