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

func (r *RoomService) HandleGuildCreate(s *discordgo.Session, g *discordgo.GuildCreate) {
	for _, c := range g.Channels {
		if c.Type == discordgo.ChannelTypeGuildVoice {
			go r.syncRoom(s, g.Guild, c)
		}
	}
}

func (r *RoomService) syncRoom(s *discordgo.Session, g *discordgo.Guild, vc *discordgo.Channel) {
	r.lock.Lock(vc.ID)
	defer r.lock.Unlock(vc.ID)

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
		var tc *discordgo.Channel
		if room.TextChannelID.Valid {
			tc, err = s.State.Channel(room.TextChannelID.String)
			if err != nil && err != discordgo.ErrStateNotFound {
				log.Printf("Failed to get text channel: %v", err)
				return
			}
		}
		if tc == nil {
			tc, err = s.GuildChannelCreateComplex(vc.GuildID, discordgo.GuildChannelCreateData{
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
			member, err := s.State.Member(gID, vs.UserID)
			if err != nil {
				return nil, err
			}
			members = append(members, member)
		}
	}
	log.Printf("Members: %v", members)
	return members, nil
}
