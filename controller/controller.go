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
	commands map[string]*Command
}

func New() *Controller {
	return &Controller{
		commands: make(map[string]*Command),
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
	c.commands[cmd.Name] = cmd
	return true
}

func (c *Controller) OverwriteGlobalCommands(s *discordgo.Session) (err error) {
	commands := make([]*discordgo.ApplicationCommand, 0, len(c.commands))
	for _, cmd := range c.commands {
		commands = append(commands, &cmd.ApplicationCommand)
	}
	_, err = s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", commands)
	return
}
