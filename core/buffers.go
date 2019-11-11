package core

import (
	"log"
)

var CMap = NewChannelMap(1000)

func BufPub(topic string, content []byte) (rv bool) {
	c := *CMap.GetOrNew(topic)
	c.Lock()
	defer c.Unlock()

	select {
	case c.c <- content:
		return true
	default:
		v := <-c.c
		// try again
		log.Printf("dropped %v\n", string(v))
		select {
		case c.c <- content:
			return true
		default:
			return false
		}
	}
}

func BufGetN(topic string, maxN int) [][]byte {
	rv := [][]byte{}
	c := *CMap.GetOrNew(topic)
	c.RLock()
	defer c.RUnlock()
	for i := 1; i <= maxN; i++ {
		select {
		case v := <-c.c:
			rv = append(rv, v)
		default:
			goto end
		}
	}

end:
	return rv
}
