package response

import "database/sql"

type ResponseRepositorier interface {
	GetResponse(guildID string, keyword string) (*Response, error)
	GetResponses(guildID string) func(yield func(*Response) bool)
	UpsertResponse(guildID string, keyword string, response string) error
	DeleteResponse(guildID string, keyword string) error
}

type ResponseRepository struct {
	db *sql.DB
}

func NewResponseRepository(db *sql.DB) *ResponseRepository {
	return &ResponseRepository{db: db}
}

func (r *ResponseRepository) GetResponse(guildID string, keyword string) (*Response, error) {
	row := r.db.QueryRow(`SELECT id, guild_id, keyword, response FROM response WHERE guild_id = $1 AND keyword = $2`, guildID, keyword)
	res := &Response{}
	err := row.Scan(&res.ID, &res.GuildID, &res.Key, &res.Response)
	return res, err
}

func (r *ResponseRepository) GetResponses(guildID string) func(yield func(*Response) bool) {
	return func(yield func(*Response) bool) {
		rows, err := r.db.Query(`SELECT id, guild_id, keyword, response FROM response WHERE guild_id = $1`, guildID)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			res := &Response{}
			err := rows.Scan(&res.ID, &res.GuildID, &res.Key, &res.Response)
			if err != nil {
				return
			}
			if !yield(res) {
				break
			}
		}
	}
}

func (r *ResponseRepository) UpsertResponse(guildID string, keyword string, response string) error {
	_, err := r.db.Exec(`INSERT INTO response (guild_id, keyword, response) VALUES ($1, $2, $3) ON CONFLICT (guild_id, keyword) DO UPDATE SET response = $3`, guildID, keyword, response)
	return err
}

func (r *ResponseRepository) DeleteResponse(guildID string, keyword string) error {
	_, err := r.db.Exec(`DELETE FROM response WHERE guild_id = $1 AND keyword = $2`, guildID, keyword)
	return err
}
