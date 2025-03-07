package translate

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

type TranslateService struct {
	c   *http.Client
	key string
}

type TranslateRequest struct {
	Text []string `json:"text"`
	To   string   `json:"target_lang"`
}

type TranslateResult struct {
	DetectedSourceLanguage string `json:"detected_source_language"`
	Text                   string `json:"text"`
}

type TranslateResponse struct {
	Translations []TranslateResult `json:"translations"`
}

func New(key string) (*TranslateService, error) {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	t := &TranslateService{c, key}
	_, err := t.Translate("Hello", "JA")
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *TranslateService) Handle(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	msg, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		return
	}

	var translated string
	var as []rune
	for _, r := range r.Emoji.Name {
		as = append(as, r-'🇦'+'A')
	}
	switch string(as) {
	case "BG":
		translated, err = t.Translate(msg.Content, "BG")
	case "CZ":
		translated, err = t.Translate(msg.Content, "CS")
	case "DK":
		translated, err = t.Translate(msg.Content, "DA")
	case "DE":
		translated, err = t.Translate(msg.Content, "DE")
	case "ER":
		translated, err = t.Translate(msg.Content, "EL")
	case "GB":
		translated, err = t.Translate(msg.Content, "EN-GB")
	case "US":
		translated, err = t.Translate(msg.Content, "EN-US")
	case "ES":
		translated, err = t.Translate(msg.Content, "ES")
	case "EE":
		translated, err = t.Translate(msg.Content, "ET")
	case "FI":
		translated, err = t.Translate(msg.Content, "FI")
	case "FR":
		translated, err = t.Translate(msg.Content, "FR")
	case "HU":
		translated, err = t.Translate(msg.Content, "HU")
	case "ID":
		translated, err = t.Translate(msg.Content, "ID")
	case "IT":
		translated, err = t.Translate(msg.Content, "IT")
	case "JP":
		translated, err = t.Translate(msg.Content, "JA")
	case "KR":
		translated, err = t.Translate(msg.Content, "KO")
	case "LV":
		translated, err = t.Translate(msg.Content, "LT")
	case "NO":
		translated, err = t.Translate(msg.Content, "NB")
	case "NL":
		translated, err = t.Translate(msg.Content, "NL")
	case "PL":
		translated, err = t.Translate(msg.Content, "PL")
	case "BR":
		translated, err = t.Translate(msg.Content, "PT-BR")
	case "PT":
		translated, err = t.Translate(msg.Content, "PT-PT")
	case "RO":
		translated, err = t.Translate(msg.Content, "RO")
	case "RU":
		translated, err = t.Translate(msg.Content, "RU")
	case "SK":
		translated, err = t.Translate(msg.Content, "SK")
	case "SI":
		translated, err = t.Translate(msg.Content, "SL")
	case "SE":
		translated, err = t.Translate(msg.Content, "SV")
	case "TR":
		translated, err = t.Translate(msg.Content, "TR")
	case "UA":
		translated, err = t.Translate(msg.Content, "UK")
	case "CN":
		translated, err = t.Translate(msg.Content, "ZH")
	}
	if err != nil {
		log.Printf("Error translating: %v", err)
		return
	}
	s.ChannelMessageSendReply(r.ChannelID, translated, msg.Reference())
}

func (t *TranslateService) Translate(text string, to string) (string, error) {

	data, _ := json.Marshal(TranslateRequest{
		Text: []string{text},
		To:   to,
	})
	req, _ := http.NewRequest("POST", "https://api-free.deepl.com/v2/translate", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "DeepL-Auth-Key "+t.key)

	res, err := t.c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var tr TranslateResponse
	err = json.NewDecoder(res.Body).Decode(&tr)
	if err != nil || len(tr.Translations) == 0 {
		return "", err
	}

	return tr.Translations[0].Text, nil
}
