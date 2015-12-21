package gwspack

import (
	"regexp"
	"sync"
)

type connpool struct {
	lock *sync.RWMutex
	pool map[string]map[*client]UserData
}

func (cp *connpool) join(c *client) (err error) {

	cp.lock.RLock()
	if v, ok := cp.pool[c.id]; !ok {
		cp.lock.RUnlock()
		m := make(map[*client]UserData)
		m[c] = c.data
		cp.lock.Lock()
		cp.pool[c.id] = m
		cp.lock.Unlock()
	} else {
		cp.lock.RUnlock()
		cp.lock.Lock()
		v[c] = c.data
		cp.lock.Unlock()
	}
	return
}

func (cp *connpool) remove(c *client) {
	cp.lock.RLock()
	if cc, ok := cp.pool[c.id]; ok {
		cp.lock.RUnlock()
		cp.lock.Lock()
		if _, ok = cc[c]; ok {
			delete(cc, c)
			close(c.send)
		}
		if len(cc) == 0 {
			delete(cp.pool, c.id)
		}
		cp.lock.Unlock()
	} else {
		cp.lock.RUnlock()
	}
	return
}
func (cp *connpool) removeById(id string) {
	cp.lock.RLock()
	if _, ok := cp.pool[id]; ok {
		cp.lock.RUnlock()
		cp.lock.Lock()
		for c, _ := range cp.pool[id] {
			delete(cp.pool[id], c)
			close(c.send)
		}
		delete(cp.pool, id)
		cp.lock.Unlock()
	} else {
		cp.lock.RUnlock()
	}

	return
}

func (cp *connpool) countById() (i int) {
	cp.lock.RLock()
	defer cp.lock.RUnlock()
	i = len(cp.pool)
	return

}

func (cp *connpool) count() (i int) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for k, _ := range cp.pool {
		for _, _ = range cp.pool[k] {
			i++
		}
	}
	return i
}

func (cp *connpool) sendTo(id string, b []byte) {
	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for c := range cp.pool[id] {
		select {
		case c.send <- b:
		default:
			close(c.send)
			delete(cp.pool[id], c)
		}
	}
	return

}

func (cp *connpool) sendAll(b []byte) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for _, clientMap := range cp.pool {
		for c := range clientMap {
			select {
			case c.send <- b:
			default:
				close(c.send)
				delete(clientMap, c)
			}
		}
	}

}

func (cp *connpool) List() (list map[string]UserData) {
	list = make(map[string]UserData)

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for k, v := range cp.pool {
		for c := range v {
			list[k] = c.data
			break
		}
	}
	return

}

func (cp *connpool) sendByRegex(vailed *regexp.Regexp, b []byte) {

	cp.lock.RLock()
	defer cp.lock.RUnlock()
	for k, clientMap := range cp.pool {
		if vailed.MatchString(k) {
			for client := range clientMap {
				client.send <- b
			}
		}
	}
}
