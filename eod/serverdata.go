package eod

import deadlock "github.com/sasha-s/go-deadlock"

func NewServerData() serverData {
	return serverData{
		lock:          &deadlock.RWMutex{},
		componentMsgs: make(map[string]componentMsg),
	}
}

func (b *EoD) setNewsChannel(channelID string, msg msg, rsp rsp) {
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_serverdata WHERE guild=? AND type=?", msg.GuildID, newsChannel)
	var count int
	err := row.Scan(&count)
	if rsp.Error(err) {
		return
	}

	if count == 1 {
		_, err = b.db.Exec("UPDATE eod_serverdata SET value1=? WHERE guild=? AND type=?", channelID, msg.GuildID, newsChannel)
		if rsp.Error(err) {
			return
		}
	} else {
		_, err = b.db.Exec("INSERT INTO eod_serverdata VALUES ( ?, ?, ?, ? )", msg.GuildID, newsChannel, channelID, 0)
		if rsp.Error(err) {
			return
		}
	}

	lock.RLock()
	dat, exists := b.dat[msg.GuildID]
	lock.RUnlock()
	if !exists {
		dat = NewServerData()
	}
	dat.newsChannel = channelID
	lock.Lock()
	b.dat[msg.GuildID] = dat
	lock.Unlock()

	rsp.Resp("Succesfully updated news channel!")
}

func (b *EoD) setVotingChannel(channelID string, msg msg, rsp rsp) {
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_serverdata WHERE guild=? AND type=?", msg.GuildID, votingChannel)
	var count int
	err := row.Scan(&count)
	if rsp.Error(err) {
		return
	}

	if count == 1 {
		_, err = b.db.Exec("UPDATE eod_serverdata SET value1=? WHERE guild=? AND type=?", channelID, msg.GuildID, votingChannel)
		if rsp.Error(err) {
			return
		}
	} else {
		_, err = b.db.Exec("INSERT INTO eod_serverdata VALUES ( ?, ?, ?, ? )", msg.GuildID, votingChannel, channelID, 0)
		if rsp.Error(err) {
			return
		}
	}

	lock.RLock()
	dat, exists := b.dat[msg.GuildID]
	lock.RUnlock()
	if !exists {
		dat = NewServerData()
	}
	dat.votingChannel = channelID
	lock.Lock()
	b.dat[msg.GuildID] = dat
	lock.Unlock()

	rsp.Resp("Succesfully updated voting channel!")
}

func (b *EoD) setVoteCount(count int, msg msg, rsp rsp) {
	if count < 0 {
		count *= -1
	}
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_serverdata WHERE guild=? AND type=?", msg.GuildID, voteCount)
	var cnt int
	err := row.Scan(&cnt)
	if rsp.Error(err) {
		return
	}

	if cnt == 1 {
		_, err = b.db.Exec("UPDATE eod_serverdata SET intval=? WHERE guild=? AND type=?", count, msg.GuildID, voteCount)
		if rsp.Error(err) {
			return
		}
	} else {
		_, err = b.db.Exec("INSERT INTO eod_serverdata VALUES ( ?, ?, ?, ? )", msg.GuildID, voteCount, "", count)
		if rsp.Error(err) {
			return
		}
	}

	lock.RLock()
	dat, exists := b.dat[msg.GuildID]
	lock.RUnlock()
	if !exists {
		dat = NewServerData()
	}
	dat.voteCount = count
	lock.Lock()
	b.dat[msg.GuildID] = dat
	lock.Unlock()

	rsp.Resp("Succesfully updated vote count!")
}

func (b *EoD) setPollCount(count int, msg msg, rsp rsp) {
	if count < 0 {
		count *= -1
	}
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_serverdata WHERE guild=? AND type=?", msg.GuildID, pollCount)
	var cnt int
	err := row.Scan(&cnt)
	if rsp.Error(err) {
		return
	}

	if cnt == 1 {
		_, err = b.db.Exec("UPDATE eod_serverdata SET intval=? WHERE guild=? AND type=?", count, msg.GuildID, pollCount)
		if rsp.Error(err) {
			return
		}
	} else {
		_, err = b.db.Exec("INSERT INTO eod_serverdata VALUES ( ?, ?, ?, ? )", msg.GuildID, pollCount, "", count)
		if rsp.Error(err) {
			return
		}
	}

	lock.RLock()
	dat, exists := b.dat[msg.GuildID]
	lock.RUnlock()
	if !exists {
		dat = NewServerData()
	}
	dat.pollCount = count
	lock.Lock()
	b.dat[msg.GuildID] = dat
	lock.Unlock()

	rsp.Resp("Succesfully updated poll count!")
}

func (b *EoD) setPlayChannel(channelID string, isPlayChannel bool, msg msg, rsp rsp) {
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_serverdata WHERE guild=? AND type=? AND value1=?", msg.GuildID, playChannel, channelID)
	var cnt int
	err := row.Scan(&cnt)
	if rsp.Error(err) {
		return
	}

	if cnt == 1 && !isPlayChannel {
		_, err = b.db.Exec("DELETE FROM eod_serverdata WHERE guild=? AND type=? AND value1=?", msg.GuildID, playChannel, channelID)
		if rsp.Error(err) {
			return
		}

		lock.RLock()
		dat, exists := b.dat[msg.GuildID]
		lock.RUnlock()
		if !exists {
			dat = NewServerData()
		}
		delete(dat.playChannels, channelID)
		lock.Lock()
		b.dat[msg.GuildID] = dat
		lock.Unlock()

		rsp.Resp("Succesfully marked channel as not a play channel.")
		return
	}

	if !isPlayChannel {
		rsp.ErrorMessage("Channel isn't play channel!")
		return
	}

	_, err = b.db.Exec("INSERT INTO eod_serverdata VALUES ( ?, ?, ?, ? )", msg.GuildID, playChannel, channelID, 0)
	if rsp.Error(err) {
		return
	}

	lock.RLock()
	dat, exists := b.dat[msg.GuildID]
	lock.RUnlock()
	if !exists {
		dat = NewServerData()
	}
	if dat.playChannels == nil {
		dat.playChannels = make(map[string]empty)
	}
	dat.playChannels[channelID] = empty{}
	lock.Lock()
	b.dat[msg.GuildID] = dat
	lock.Unlock()

	rsp.Resp("Succesfully marked channel as play channel!")
}

func (b *EoD) setModRole(roleID string, msg msg, rsp rsp) {
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_serverdata WHERE guild=? AND type=?", msg.GuildID, modRole)
	var count int
	err := row.Scan(&count)
	if rsp.Error(err) {
		return
	}

	if count == 1 {
		_, err = b.db.Exec("UPDATE eod_serverdata SET value1=? WHERE guild=? AND type=?", roleID, msg.GuildID, modRole)
		if rsp.Error(err) {
			return
		}
	} else {
		_, err = b.db.Exec("INSERT INTO eod_serverdata VALUES ( ?, ?, ?, ? )", msg.GuildID, modRole, roleID, 0)
		if rsp.Error(err) {
			return
		}
	}

	lock.RLock()
	dat, exists := b.dat[msg.GuildID]
	lock.RUnlock()
	if !exists {
		dat = NewServerData()
	}
	dat.modRole = roleID
	lock.Lock()
	b.dat[msg.GuildID] = dat
	lock.Unlock()

	rsp.Resp("Succesfully updated mod role!")
}
