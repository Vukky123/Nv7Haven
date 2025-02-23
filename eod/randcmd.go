package eod

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var ideaCmp = discordgo.ActionsRow{
	Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "New Idea",
			Style:    discordgo.SuccessButton,
			CustomID: "idea",
		},
	},
}

type ideaComponent struct {
	catName string
	hasCat  bool
	count   int
	b       *EoD
}

func (c *ideaComponent) handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	res, suc := c.b.genIdea(c.count, c.catName, c.hasCat, i.GuildID, i.Member.User.ID)
	if !suc {
		res += " " + redCircle
	}
	err := c.b.dg.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    res,
			Components: []discordgo.MessageComponent{ideaCmp},
		},
	})
	if err != nil {
		log.Println("Failed to send message:", err)
	}
}

func (b *EoD) genIdea(count int, catName string, hasCat bool, guild string, author string) (string, bool) {
	if count > maxComboLength {
		return fmt.Sprintf("You can only combine up to %d elements!", maxComboLength), false
	}

	if count < 2 {
		return "You must combine at least 2 elements!", false
	}

	lock.RLock()
	dat, exists := b.dat[guild]
	lock.RUnlock()
	if !exists {
		return "Guild not found", false
	}

	dat.lock.RLock()
	inv, exists := dat.invCache[author]
	dat.lock.RUnlock()
	if !exists {
		return "You don't have an inventory!", false
	}

	els := inv
	if hasCat {
		cat, exists := dat.catCache[strings.ToLower(catName)]
		if !exists {
			return fmt.Sprintf("Category **%s** doesn't exist!", catName), false
		}
		els = make(map[string]empty)

		for el := range cat.Elements {
			l := strings.ToLower(el)
			_, exists := inv[l]
			if exists {
				els[l] = empty{}
			}
		}

		if len(els) == 0 {
			return fmt.Sprintf("You don't have any elements in category **%s**!", cat.Name), false
		}
	}

	var elem3 string
	cont := true
	var elems []string
	tries := 0
	for cont {
		elems = make([]string, count)
		for i := range elems {
			cnt := rand.Intn(len(els))
			j := 0
			for k := range els {
				if j == cnt {
					elems[i] = k
					break
				}
				j++
			}
		}

		query := "SELECT elem3 FROM eod_combos WHERE elems LIKE ? AND guild=?"
		els := elems2txt(elems)
		if isASCII(els) {
			query = "SELECT elem3 FROM eod_combos WHERE CONVERT(elems USING utf8mb4) LIKE CONVERT(? USING utf8mb4) AND guild=CONVERT(? USING utf8mb4) COLLATE utf8mb4_general_ci"
		}
		row := b.db.QueryRow(query, els, guild)
		err := row.Scan(&elem3)
		if err != nil {
			cont = false
		}
		tries++

		if tries > 10 {
			return "Couldn't find a random unused combination, maybe try again later?", false
		}
	}

	text := ""
	for i, el := range elems {
		dat.lock.RLock()
		text += dat.elemCache[strings.ToLower(el)].Name
		dat.lock.RUnlock()
		if i != len(elems)-1 {
			text += " + "
		}
	}

	if dat.combCache == nil {
		dat.combCache = make(map[string]comb)
	}
	dat.lock.Lock()
	dat.combCache[author] = comb{
		elems: elems,
		elem3: "",
	}
	dat.lock.Unlock()

	lock.Lock()
	b.dat[guild] = dat
	lock.Unlock()

	return fmt.Sprintf("Your random unused combination is... **%s**\n 	Suggest it by typing **/suggest**", text), true
}
func (b *EoD) ideaCmd(count int, catName string, hasCat bool, m msg, rsp rsp) {
	res, suc := b.genIdea(count, catName, hasCat, m.GuildID, m.Author.ID)
	if !suc {
		rsp.ErrorMessage(res)
		return
	}
	rsp.Acknowledge()

	lock.Lock()
	dat, exists := b.dat[m.GuildID]
	lock.Unlock()
	if !exists {
		rsp.ErrorMessage("Guild not found")
		return
	}

	id := rsp.Message(res, ideaCmp)

	dat.lock.Lock()
	dat.componentMsgs[id] = &ideaComponent{
		catName: catName,
		count:   count,
		hasCat:  hasCat,
		b:       b,
	}
	dat.lock.Unlock()

	lock.Lock()
	b.dat[m.GuildID] = dat
	lock.Unlock()
}
