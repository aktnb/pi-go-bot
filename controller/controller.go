package controller

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type InteractionHandler interface {
	HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

// スラッシュコマンドの構造体
type Command struct {
	discordgo.ApplicationCommand
	Execute func(s *discordgo.Session, i *discordgo.InteractionCreate) error
}

func (c *Command) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return c.Execute(s, i)
}

type Controller struct {
	commands map[string]InteractionHandler
}

func New() *Controller {
	return &Controller{
		commands: make(map[string]InteractionHandler),
	}
}

func (c *Controller) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if command, ok := c.commands[data.Name]; ok {
		err := command.HandleInteraction(s, i)
		if err != nil {
			log.Printf("Error executing command %s: %v", data.Name, err)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "An error occurred while executing the command",
					Flags:   discordgo.MessageFlagsEphemeral,
				}},
			)
		}
		return
	}
	log.Printf("Command %s not found", data.Name)
}

func (c *Controller) AddGlobalCommand(s *discordgo.Session, cmd *Command) bool {
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", &cmd.ApplicationCommand)
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
		return false
	}

	log.Printf("Command %s added", ccmd.Name)
	cmd.ApplicationCommand = *ccmd
	c.commands[cmd.Name] = cmd
	return true
}
