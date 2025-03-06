package room

import (
	"database/sql"
	"log"

	"github.com/aktnb/pi-go-bot/pkg/keylock"
	"github.com/bwmarrin/discordgo"
)

type Room struct {
	ID             int64
	VoiceChannelID string
	TextChannelID  sql.NullString
}

type RoomService struct {
	lock           keylock.KeyLock
	roomRepository RoomRepositorier
}

func New(db *sql.DB) *RoomService {
	return &RoomService{
		*keylock.New(),
		NewRoomRepository(db),
	}
}

func (r *RoomService) HandleVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	if v.BeforeUpdate != nil && v.VoiceState != nil && v.BeforeUpdate.ChannelID == v.VoiceState.ChannelID {
		return
	}

	if v.BeforeUpdate != nil && v.BeforeUpdate.ChannelID != "" {
		//	退出
		go r.leaveRoom(s, v.BeforeUpdate)
	}

	if v.VoiceState != nil && v.VoiceState.ChannelID != "" {
		//	入室
		go r.joinRoom(s, v.VoiceState)
	}
}

func (r *RoomService) joinRoom(s *discordgo.Session, v *discordgo.VoiceState) {
	r.lock.Lock(v.ChannelID)
	defer r.lock.Unlock(v.ChannelID)

	vc, err := s.State.Channel(v.ChannelID)
	if err != nil {
		log.Printf("Failed to get voice channel: %v", err)
		return
	}

	room, err := r.roomRepository.GetByVoiceChannelID(v.ChannelID)
	if err != nil {
		log.Printf("Failed to get room: %v", err)
		return
	}
	if room == nil {
		room = &Room{
			VoiceChannelID: v.ChannelID,
		}
	}

	var tc *discordgo.Channel
	if room.TextChannelID.Valid {
		tc, err = s.State.Channel(room.TextChannelID.String)
		if err != nil && err != discordgo.ErrStateNotFound {
			log.Printf("Failed to get text channel: %v", err)
			return
		}
	}
	if tc == nil {
		tc, err = s.GuildChannelCreateComplex(v.GuildID, discordgo.GuildChannelCreateData{
			Name:     vc.Name + "-text",
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: vc.ParentID,
			Position: vc.Position,
			PermissionOverwrites: []*discordgo.PermissionOverwrite{
				{
					ID:    s.State.User.ID,
					Type:  discordgo.PermissionOverwriteTypeMember,
					Allow: discordgo.PermissionViewChannel,
				},
				{
					ID:   v.GuildID,
					Type: discordgo.PermissionOverwriteTypeRole,
					Deny: discordgo.PermissionViewChannel,
				},
			},
		})
		if err != nil {
			log.Printf("Failed to create text channel: %v", err)
			return
		}
		room.TextChannelID = sql.NullString{String: tc.ID, Valid: true}
	}

	err = s.ChannelPermissionSet(tc.ID, v.UserID, discordgo.PermissionOverwriteTypeMember, discordgo.PermissionViewChannel, 0)
	if err != nil {
		log.Printf("Failed to set channel permission: %v", err)
		return
	}

	err = r.roomRepository.Upsert(room)
	if err != nil {
		log.Printf("Failed to upsert room: %v", err)
		return
	}
}

func (r *RoomService) leaveRoom(s *discordgo.Session, v *discordgo.VoiceState) {
	r.lock.Lock(v.ChannelID)
	defer r.lock.Unlock(v.ChannelID)

	room, err := r.roomRepository.GetByVoiceChannelID(v.ChannelID)
	if err != nil {
		log.Printf("Failed to get room: %v", err)
		return
	}
	if room == nil {
		return
	}

	if !room.TextChannelID.Valid {
		return
	}

	err = s.ChannelPermissionSet(room.TextChannelID.String, v.UserID, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionViewChannel)
	if err != nil {
		log.Printf("Failed to set channel permission: %v", err)
		return
	}

	vcMembers, err := r.getVoiceChannelMembers(s, v.GuildID, v.ChannelID)
	if err != nil {
		log.Printf("Failed to get voice channel members: %v", err)
		return
	}

	if len(vcMembers) == 0 {
		_, err := s.ChannelDelete(room.TextChannelID.String)
		if err != nil {
			log.Printf("Failed to delete text channel: %v", err)
			return
		}
		room.TextChannelID = sql.NullString{String: "", Valid: false}
	}

	err = r.roomRepository.Upsert(room)
	if err != nil {
		log.Printf("Failed to upsert room: %v", err)
		return
	}
}

func (r *RoomService) getVoiceChannelMembers(s *discordgo.Session, gID string, vcID string) ([]*discordgo.Member, error) {
	//	Todo: 1000 件以上のメンバーを取得するために、複数回に分けて取得する
	g, err := s.State.Guild(gID)
	if err != nil {
		return nil, err
	}

	var members []*discordgo.Member
	for _, vs := range g.VoiceStates {
		if vs.ChannelID == vcID {
			members = append(members, vs.Member)
		}
	}
	log.Printf("Members: %v", members)
	return members, nil
}
