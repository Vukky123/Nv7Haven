package eod

import (
	"encoding/json"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (b *EoD) init() {
	for _, v := range commands {
		go func(val *discordgo.ApplicationCommand) {
			_, err := b.dg.ApplicationCommandCreate(clientID, "819077688371314718", val)
			if err != nil {
				panic(err)
			}
		}(v)
	}
	b.dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

	res, err := b.db.Query("SELECT * FROM eod_serverdata WHERE 1")
	if err != nil {
		panic(err)
	}
	defer res.Close()

	var guild string
	var kind serverDataType
	var value1 string
	var intval int
	for res.Next() {
		err = res.Scan(&guild, &kind, &value1, &intval)
		if err != nil {
			panic(err)
		}

		switch kind {
		case newsChannel:
			lock.RLock()
			dat, exists := b.dat[guild]
			lock.RUnlock()
			if !exists {
				dat = serverData{}
			}
			dat.newsChannel = value1
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()

		case playChannel:
			lock.RLock()
			dat, exists := b.dat[guild]
			lock.RUnlock()
			if !exists {
				dat = serverData{}
			}
			if dat.playChannels == nil {
				dat.playChannels = make(map[string]empty)
			}
			dat.playChannels[value1] = empty{}
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()

		case votingChannel:
			lock.RLock()
			dat, exists := b.dat[guild]
			lock.RUnlock()
			if !exists {
				dat = serverData{}
			}
			dat.votingChannel = value1
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()

		case voteCount:
			lock.RLock()
			dat, exists := b.dat[guild]
			lock.RUnlock()
			if !exists {
				dat = serverData{}
			}
			dat.voteCount = intval
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()
		}
	}

	elems, err := b.db.Query("SELECT * FROM eod_elements WHERE 1")
	if err != nil {
		panic(err)
	}
	defer elems.Close()
	elem := element{}
	var createdon int64
	var parent1 string
	var parent2 string
	for elems.Next() {
		err = elems.Scan(&elem.Name, &elem.Category, &elem.Guild, &elem.Comment, &elem.Creator, &createdon, &parent1, &parent2, &elem.Complexity)
		if err != nil {
			return
		}
		elem.CreatedOn = time.Unix(createdon, 0)
		if parent1 != "" && parent2 != "" {
			elem.Parents = []string{parent1, parent2}
		} else {
			elem.Parents = make([]string, 0)
		}

		lock.RLock()
		dat := b.dat[elem.Guild]
		lock.RUnlock()
		if dat.elemCache == nil {
			dat.elemCache = make(map[string]element)
		}
		dat.elemCache[elem.Name] = elem
		lock.Lock()
		b.dat[elem.Guild] = dat
		lock.Unlock()
	}

	invs, err := b.db.Query("SELECT guild, user, inv FROM eod_inv WHERE 1")
	if err != nil {
		panic(err)
	}
	defer invs.Close()
	var invDat string
	var user string
	var inv map[string]empty
	for invs.Next() {
		inv = make(map[string]empty, 0)
		err = invs.Scan(&guild, &user, &invDat)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal([]byte(invDat), &inv)
		if err != nil {
			panic(err)
		}
		lock.RLock()
		dat := b.dat[guild]
		lock.RUnlock()
		dat.invCache[user] = inv
		lock.Lock()
		b.dat[guild] = dat
		lock.Unlock()
	}

	polls, err := b.db.Query("SELECT * FROM eod_polls WHERE 1")
	if err != nil {
		panic(err)
	}
	defer polls.Close()
	var po poll
	for polls.Next() {
		err = polls.Scan(guild, &po.Channel, &po.Message, &po.Kind, &po.Value1, &po.Value2, &po.Value3, &po.Value4)
		if err != nil {
			panic(err)
		}

		lock.RLock()
		dat := b.dat[guild]
		lock.RUnlock()
		if dat.polls == nil {
			dat.polls = make([]poll, 0)
		}
		dat.polls = append(dat.polls, po)
		lock.Lock()
		b.dat[guild] = dat
		lock.Unlock()

		b.dg.ChannelMessageDelete(po.Channel, po.Message)
		err = b.createPoll(po)
		if err != nil {
			panic(err)
		}
	}
}
