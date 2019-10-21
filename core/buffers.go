package core

import (
	"log"
)

var CMap = NewChannelMap(10000)

func BufPub(topic string, content []byte) (rv bool) {
	c := *CMap.GetOrNew(topic)
	select {
	case c <- content:
		return true
	default:
		v := <-c
		// try again
		log.Printf("dropped %v\n", string(v))
		select {
		case c <- content:
			return true
		default:
			return false
		}
	}
}

func BufGetN(topic string, maxN int) [][]byte {
	rv := [][]byte{}
	c := *CMap.GetOrNew(topic)
	for i := 1; i <= maxN; i++ {
		select {
		case v := <-c:
			rv = append(rv, v)
		default:
			goto end
		}
	}

end:
	return rv
}
