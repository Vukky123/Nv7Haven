package discord

import (
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var warnMatch = regexp.MustCompile(`warn <@!?\d+> (.+)`)

type warning struct {
	Mod  string //ID
	Text string
	Date int64
}

func (b *Bot) mod(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "warn") {
		if !(len(m.Mentions) > 0) {
			s.ChannelMessageSend(m.ChannelID, "You need to mention the person you are going to warn!")
			return
		}

		groups := warnMatch.FindAllStringSubmatch(m.Content, -1)
		if len(groups) < 1 {
			s.ChannelMessageSend(m.ChannelID, "Does not match format `warn @user <warning text>`")
			return
		}
		messageCont := groups[0]
		if len(messageCont) < 2 {
			s.ChannelMessageSend(m.ChannelID, "Does not match format `warn @user <warning text>`")
			return
		}
		message := messageCont[1]

		b.checkuserwithid(m, m.Mentions[0].ID)

		if b.isMod(m, m.Author.ID) {
			warn := warning{
				Mod:  m.Author.ID,
				Text: message,
				Date: time.Now().Unix(),
			}
			user, suc := b.getuser(m, m.Mentions[0].ID)
			if !suc {
				return
			}
			var existing []warning
			_, exists := user.Metadata["warns"]
			if !exists {
				existing = make([]warning, 0)
			} else {
				existing = user.Metadata["warns"].([]warning)
			}
			existing = append(existing, warn)
			user.Metadata["warns"] = existing
			suc = b.updateuser(m, user)
			if !suc {
				return
			}
			s.ChannelMessageSend(m.ChannelID, `Successfully warned user.`)
			return
		}
		s.ChannelMessageSend(m.ChannelID, `You need to have permission "Administrator" to use this command.`)
		return
	}
}
