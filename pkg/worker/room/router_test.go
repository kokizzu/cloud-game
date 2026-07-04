package room

import (
	"sync"
	"testing"
	"time"

	"github.com/giongto35/cloud-game/v3/pkg/com"
)

type sKey string

func (s sKey) String() string { return string(s) }

type tSession struct {
	id        sKey
	connected bool
}

func (t *tSession) SendAudio([]byte, time.Duration) {}
func (t *tSession) SendVideo([]byte, time.Duration) {}
func (t *tSession) SendData([]byte)                 {}
func (t *tSession) Connect()                        { t.connected = true }
func (t *tSession) Disconnect()                     { t.connected = false }
func (t *tSession) Id() sKey                        { return t.id }

type lookMap struct {
	com.NetMap[sKey, *tSession]
	prev com.NetMap[sKey, *tSession] // we could use pointers in the original :3
}

func (l *lookMap) Reset() {
	l.prev = com.NewNetMap[sKey, *tSession]()
	for s := range l.Map.Values() {
		l.prev.Add(s)
	}
	l.NetMap.Reset()
}

func TestRouter(t *testing.T) {
	router := newTestRouter()

	var r *Room[*tSession]

	router.SetRoom(&Room[*tSession]{id: "test001"})
	room := router.FindRoom("test001")
	if room == nil {
		t.Errorf("no room, but should be")
	}
	router.SetRoom(r)
	room = router.FindRoom("x")
	if room != nil {
		t.Errorf("a room, but should not be")
	}
	router.SetRoom(nil)
	router.Close()
}

func TestRouterReset(t *testing.T) {
	u := lookMap{NetMap: com.NewNetMap[sKey, *tSession]()}
	router := Router[*tSession]{users: &u}

	router.AddUser(&tSession{id: "1", connected: true})
	router.AddUser(&tSession{id: "2", connected: false})
	router.AddUser(&tSession{id: "3", connected: true})

	router.Reset()

	disconnected := true
	for u := range u.prev.Values() {
		disconnected = disconnected && !u.connected
	}
	if !disconnected {
		t.Errorf("not all users were disconnected, but should")
	}
	if !router.Users().Empty() {
		t.Errorf("has users after reset, but should not")
	}
}

func TestRouterRemove(t *testing.T) {
	router := newTestRouter()
	router.room = &Room[*tSession]{}
	u1 := &tSession{id: "u1"}
	u2 := &tSession{id: "u2"}
	router.AddUser(u1)
	router.AddUser(u2)

	closed := false
	router.room.HandleClose = func() { closed = true }

	router.Remove(u1)
	if router.Users().Empty() {
		t.Error("should have user u2")
	}
	if closed {
		t.Error("room should not close while wiht users")
	}

	router.Remove(u2)
	if !router.Users().Empty() {
		t.Error("should be empty after last user removed")
	}
	if !closed {
		t.Error("room should close when last user leaves")
	}
}

func TestRouterRemove_Race(t *testing.T) {
	router := newTestRouter()
	router.room = &Room[*tSession]{}

	var emptyAtClose bool
	router.room.HandleClose = func() {
		emptyAtClose = router.Users().Empty()
	}

	var wg sync.WaitGroup
	n := 5
	wg.Add(n)

	for range n {
		go func() {
			defer wg.Done()
			for range 100 {
				u := &tSession{id: "u"}
				router.AddUser(u)
				router.Remove(u)
			}
		}()
	}

	wg.Wait()

	if closed := router.Room() == nil; closed && !emptyAtClose {
		t.Error("room closed while users still present")
	}
}

func newTestRouter() *Router[*tSession] {
	u := com.NewNetMap[sKey, *tSession]()
	return &Router[*tSession]{users: &u}
}
