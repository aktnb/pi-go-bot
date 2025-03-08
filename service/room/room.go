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
		go r.syncRoom(s, v.BeforeUpdate.ChannelID)
	}

	if v.VoiceState != nil && v.VoiceState.ChannelID != "" {
		//	入室
		go r.syncRoom(s, v.VoiceState.ChannelID)
	}
}

func (r *RoomService) HandleGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	for _, c := range g.Channels {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			go r.syncRoom(s, c.ID)
		}
	}
}

func (r *RoomService) syncRoom(s *discordgo.Session, vID string) {
	r.lock.Lock(vID)
	defer r.lock.Unlock(vID)

	vc, err := s.State.Channel(vID)
	if err != nil {
		log.Printf("Failed to get voice channel: %v", err)
		return
	}
	g, err := s.State.Guild(vc.GuildID)
	if err != nil {
		log.Printf("Failed to get guild: %v", err)
		return
	}

	log.Printf("Sync room: %v\n", vc.Name)

	vcMembers, err := r.getVoiceChannelMembers(s, vc.GuildID, vc.ID)
	if err != nil {
		log.Printf("Failed to get voice channel members: %v", err)
		return
	}
	room, err := r.roomRepository.GetByVoiceChannelID(vc.ID)
	if err != nil {
		log.Printf("Failed to get room: %v", err)
		return
	}
	if room == nil {
		room = &Room{
			VoiceChannelID: vc.ID,
		}
	}

	if len(vcMembers) == 0 {
		if room.TextChannelID.Valid {
			_, err := s.ChannelDelete(room.TextChannelID.String)
			if err != nil {
				log.Printf("Failed to delete text channel: %v", err)
				return
			}
			room.TextChannelID = sql.NullString{String: "", Valid: false}
		}
	} else {
		tc, err := r.PrepareTextChannel(s, vc, room)
		if err != nil {
			log.Printf("Failed to prepare text channel: %v", err)
			return
		}

		for _, m := range g.Members {
			if m.User.ID == s.State.User.ID {
				continue
			}
			var found bool
			for _, vm := range vcMembers {
				if m.User.ID == vm.User.ID {
					found = true
					break
				}
			}
			permission, err := s.State.UserChannelPermissions(m.User.ID, tc.ID)
			if err != nil && err != discordgo.ErrStateNotFound {
				log.Printf("Failed to get user channel permissions: %v", err)
				return
			}
			if !found && permission&discordgo.PermissionViewChannel != 0 {
				err = s.ChannelPermissionSet(tc.ID, m.User.ID, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionViewChannel)
				if err != nil {
					log.Printf("Failed to set channel permission: %v", err)
					return
				}
			} else if found && permission&discordgo.PermissionViewChannel == 0 {
				err = s.ChannelPermissionSet(tc.ID, m.User.ID, discordgo.PermissionOverwriteTypeMember, discordgo.PermissionViewChannel, 0)
				if err != nil {
					log.Printf("Failed to set channel permission: %v", err)
					return
				}
			}
		}
	}
	r.roomRepository.Upsert(room)
}

func (r *RoomService) getVoiceChannelMembers(s *discordgo.Session, gID string, vcID string) ([]*discordgo.Member, error) {
	g, err := s.State.Guild(gID)
	if err != nil {
		return nil, err
	}

	var members []*discordgo.Member
	for _, vs := range g.VoiceStates {
		if vs.ChannelID == vcID {
			member, err := s.State.Member(gID, vs.UserID)
			if err != nil {
				return nil, err
			}
			members = append(members, member)
		}
	}
	return members, nil
}

func (r *RoomService) PrepareTextChannel(s *discordgo.Session, vc *discordgo.Channel, room *Room) (tc *discordgo.Channel, err error) {
	if room.TextChannelID.Valid {
		tc, err = s.State.Channel(room.TextChannelID.String)
		if err != nil && err != discordgo.ErrStateNotFound {
			log.Printf("Failed to get text channel: %v", err)
			room.TextChannelID = sql.NullString{String: "", Valid: false}
			return
		}
	}
	if tc == nil {
		tc, err = s.GuildChannelCreateComplex(vc.GuildID, discordgo.GuildChannelCreateData{
			Name:     "通話用テキストチャンネル",
			Type:     discordgo.ChannelTypeGuildText,
			ParentID: vc.ParentID,
			Position: 0,
			PermissionOverwrites: []*discordgo.PermissionOverwrite{
				{
					ID:    s.State.User.ID,
					Type:  discordgo.PermissionOverwriteTypeMember,
					Allow: discordgo.PermissionViewChannel,
				},
				{
					ID:   vc.GuildID,
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
	return
}
