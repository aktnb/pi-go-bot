package room

import "database/sql"

type RoomRepositorier interface {
	GetByVoiceChannelID(voiceChannelID string) (*Room, error)
	Upsert(room *Room) error
}

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) RoomRepositorier {
	return &RoomRepository{
		db: db,
	}
}

func (r *RoomRepository) GetByVoiceChannelID(voiceChannelID string) (*Room, error) {
	var room Room
	row := r.db.QueryRow("SELECT id, voicechannel_id, textchannel_id FROM room WHERE voicechannel_id = $1", voiceChannelID)
	if err := row.Scan(&room.ID, &room.VoiceChannelID, &room.TextChannelID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) Upsert(room *Room) error {
	_, err := r.db.Exec("INSERT INTO room (voicechannel_id, textchannel_id) VALUES ($1, $2) ON CONFLICT (voicechannel_id) DO UPDATE SET textchannel_id = $2", room.VoiceChannelID, room.TextChannelID)
	return err
}
