package room

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Room struct {
	ID             int64
	VoiceChannelID string
	TextChannelID  string
}

type RoomService struct {
}

func New() *RoomService {
	return &RoomService{}
}

func (r *RoomService) HandleVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	// ここに処理を書く
	fmt.Println("VoiceStateUpdate")
}
