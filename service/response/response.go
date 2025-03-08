package response

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
)

type Response struct {
	ID       int
	GuildID  string
	Key      string
	Response string
}

type ResponseService struct {
	responseRepository ResponseRepositorier
}

func New(db *sql.DB) *ResponseService {
	return &ResponseService{
		responseRepository: NewResponseRepository(db),
	}
}

func (s *ResponseService) HandleMessageCreate(se *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	response, err := s.GetResponse(m.GuildID, m.Content)
	if err != nil {
		return
	}

	if response == "" {
		return
	}

	se.ChannelMessageSend(m.ChannelID, response)
}

func (s *ResponseService) SetResponse(guildID string, keyword string, response string) error {
	return s.responseRepository.UpsertResponse(guildID, keyword, response)
}

func (s *ResponseService) DeleteResponse(guildID string, keyword string) error {
	return s.responseRepository.DeleteResponse(guildID, keyword)
}

func (s *ResponseService) GetResponse(guildID string, keyword string) (string, error) {
	res, err := s.responseRepository.GetResponse(guildID, keyword)
	if err != nil {
		return "", err
	}
	return res.Response, nil
}

func (s *ResponseService) GetKeys(guildID string) func(yield func(string) bool) {
	return func(yield func(string) bool) {
		for res := range s.responseRepository.GetResponses(guildID) {
			if !yield(res.Key) {
				return
			}
		}
	}
}
