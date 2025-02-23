package eod

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

func (b *EoD) init() {
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
				dat = NewServerData()
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
				dat = NewServerData()
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
				dat = NewServerData()
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
				dat = NewServerData()
			}
			dat.voteCount = intval
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()

		case pollCount:
			lock.RLock()
			dat, exists := b.dat[guild]
			lock.RUnlock()
			if !exists {
				dat = NewServerData()
			}
			dat.pollCount = intval
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()

		case modRole:
			lock.RLock()
			dat, exists := b.dat[guild]
			lock.RUnlock()
			if !exists {
				dat = NewServerData()
			}
			dat.modRole = value1
			lock.Lock()
			b.dat[guild] = dat
			lock.Unlock()
		}
	}

	//elems, err := b.db.Query("SELECT * FROM eod_elements ORDER BY createdon ASC") // Do after nov 21

	var cnt int
	err = b.db.QueryRow("SELECT COUNT(1) FROM eod_elements").Scan(&cnt)
	if err != nil {
		panic(err)
	}

	bar := progressbar.New(cnt)

	elems, err := b.db.Query("SELECT name, image, guild, comment, creator, createdon, parents, complexity, difficulty, usedin FROM `eod_elements` ORDER BY (CASE WHEN createdon=1637536881 THEN 1605988759 ELSE createdon END) ASC")
	if err != nil {
		panic(err)
	}
	defer elems.Close()
	elem := element{}
	var createdon int64
	var parentDat string
	for elems.Next() {
		err = elems.Scan(&elem.Name, &elem.Image, &elem.Guild, &elem.Comment, &elem.Creator, &createdon, &parentDat, &elem.Complexity, &elem.Difficulty, &elem.UsedIn)
		if err != nil {
			return
		}
		elem.CreatedOn = time.Unix(createdon, 0)

		if len(parentDat) == 0 {
			elem.Parents = make([]string, 0)
		} else {
			elem.Parents = strings.Split(parentDat, "+")
		}

		lock.RLock()
		dat := b.dat[elem.Guild]
		lock.RUnlock()
		if dat.elemCache == nil {
			dat.elemCache = make(map[string]element)
		}
		elem.ID = len(dat.elemCache) + 1
		dat.elemCache[strings.ToLower(elem.Name)] = elem
		lock.Lock()
		b.dat[elem.Guild] = dat
		lock.Unlock()

		bar.Add(1)
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
		inv = make(map[string]empty)
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
		if dat.invCache == nil {
			dat.invCache = make(map[string]map[string]empty)
		}
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
		var jsondat string
		err = polls.Scan(&guild, &po.Channel, &po.Message, &po.Kind, &po.Value1, &po.Value2, &po.Value3, &po.Value4, &jsondat)
		if err != nil {
			panic(err)
		}
		po.Guild = guild
		err = json.Unmarshal([]byte(jsondat), &po.Data)
		if err != nil {
			panic(err)
		}

		_, err = b.db.Exec("DELETE FROM eod_polls WHERE guild=? AND channel=? AND message=?", po.Guild, po.Channel, po.Message)
		if err != nil {
			panic(err)
		}

		b.dg.ChannelMessageDelete(po.Channel, po.Message)
		err = b.createPoll(po)
		if err != nil {
			fmt.Println(err)
		}
	}

	lock.RLock()
	for k, dat := range b.dat {
		hasChanged := false
		if dat.invCache == nil {
			dat.invCache = make(map[string]map[string]empty)
			hasChanged = true
		}
		if hasChanged {
			lock.RUnlock()
			lock.Lock()
			b.dat[k] = dat
			lock.Unlock()
			lock.RLock()
		}
	}
	lock.RUnlock()

	cats, err := b.db.Query("SELECT * FROM eod_categories")
	if err != nil {
		panic(err)
	}
	defer cats.Close()
	var elemDat string
	cat := category{}
	for cats.Next() {
		err = cats.Scan(&guild, &cat.Name, &elemDat, &cat.Image)
		if err != nil {
			return
		}

		cat.Guild = guild

		lock.RLock()
		dat := b.dat[guild]
		lock.RUnlock()
		if dat.catCache == nil {
			dat.catCache = make(map[string]category)
		}

		cat.Elements = make(map[string]empty)
		err := json.Unmarshal([]byte(elemDat), &cat.Elements)
		if err != nil {
			panic(err)
		}

		dat.catCache[strings.ToLower(cat.Name)] = cat
		lock.Lock()
		b.dat[guild] = dat
		lock.Unlock()
	}

	b.initHandlers()

	// Start stats saving
	go func() {
		b.saveStats()
		for {
			time.Sleep(time.Minute * 30)
			b.saveStats()
		}
	}()
}
