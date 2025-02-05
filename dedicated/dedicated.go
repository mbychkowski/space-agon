// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	// "encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"net"
	"os"
	"sync"
	"time"

	agonesSdk "agones.dev/agones/pkg/sdk"
	agones "agones.dev/agones/sdks/go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/mbychkowski/space-agon/game"
	"github.com/mbychkowski/space-agon/game/pb"
	"github.com/mbychkowski/space-agon/game/protostream"
	"golang.org/x/net/websocket"
)

type GameEvent struct {
	EventID   string
	PlayerID  int64
	EventType string
	Timestamp int64
	Data      interface{}
}

const (
	listApiHost = "listener.default.svc.cluster.local"
	listApiPort = "7777"
)

var (
	addr = flag.String("addr", listApiHost+":"+listApiPort, "the address to connect to")
	name = flag.String("name", "MikeB!", "Name to greet")
)

func main() {

	playerConnected, playerDisconnected := startAgones()

	http.Handle("/connect/", newDedicated(playerConnected, playerDisconnected))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Println("Starting dedicated server!!")
	log.Fatal(http.ListenAndServe(":2156", nil))
}

type dedicated struct {
	g *game.Game

	nextCid chan int64

	mr *memoRouter

	playerConnected    func()
	playerDisconnected func()
}

func newDedicated(playerConnected func(), playerDisconnected func()) websocket.Handler {
	d := &dedicated{
		g:                  game.NewGame(),
		nextCid:            make(chan int64, 1),
		mr:                 newMemoRouter(),
		playerConnected:    playerConnected,
		playerDisconnected: playerDisconnected,
	}
	inp := game.NewInput()
	inp.IsRendered = false
	inp.IsPlayer = false
	inp.IsHost = true

	d.nextCid <- 1

	go func() {
		toSend, receive := d.mr.connect(0)

		last := time.Now()
		for t := range time.Tick(time.Second / 60) {
			select {
			case inp.Memos = <-toSend:
			default:
				inp.Memos = nil
			}

			inp.Dt = float32(t.Sub(last).Seconds())
			last = t
			d.g.Step(inp)

			receive(inp.MemosOut)
			inp.MemosOut = nil
		}
	}()

	return d.Handler
}

func (d *dedicated) Handler(c *websocket.Conn) {
	c.PayloadType = 2 // Sets sent payloads to binary

	d.playerConnected()
	defer d.playerDisconnected()

	ctx, cancel := context.WithCancel(context.Background())

	cid := <-d.nextCid
	d.nextCid <- cid + 1

	log.Println("Client ID", cid, d.nextCid)

	toSend, recieve := d.mr.connect(cid)
	defer d.mr.disconnect(cid)

	stream := protostream.NewProtoStream(c)

	go func() {
		defer cancel()
		err := stream.Send(&pb.ClientInitialize{Cid: cid})
		if err != nil {
			log.Printf("Client %d had send clientInitialize error %v", cid, err)
			return
		}

		for {
			select {
			case memos := <-toSend:
				err := stream.Send(&pb.Memos{Memos: memos})
				if err != nil {
					log.Printf("Client %d had send memos error %v", cid, err)
					return
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	// conn := open(listApiHost+":"+listApiPort)

	go func() {
		// ge := &GameEvent{
		// 	EventID: "1234",
		// 	PlayerID: cid,
		// 	EventType: "NoEvent",
		// 	Timestamp: time.Now().Unix(),
		// 	Data: "test event data...",
		// }
		// writeEvent(ge, conn)

		defer cancel()

		for {

			memos := &pb.Memos{}

			// log.Println("mem", memos)

			err := stream.Recv(memos)
			if err != nil {
				log.Printf("Client %d had read/decode error %v", cid, err)
				return
			}
			recieve(memos.Memos)
		}
	}()

	<-ctx.Done()
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

func combineToSend(c chan []*pb.Memo, memos []*pb.Memo) {
	select {
	case previousMemos := <-c:
		previousMemos = append(previousMemos, memos...)
		c <- previousMemos
	case c <- memos:
	}
}

type memoRouter struct {
	incoming     chan []*pb.Memo
	outgoing     map[int64]chan []*pb.Memo
	outgoingLock sync.Mutex
	createMemos  map[uint64]*pb.Memo
}

func newMemoRouter() *memoRouter {
	mr := &memoRouter{
		incoming: make(chan []*pb.Memo, 1),
		outgoing: make(map[int64]chan []*pb.Memo),

		createMemos: make(map[uint64]*pb.Memo),
	}

	go func() {
		for memos := range mr.incoming {
			mr.outgoingLock.Lock()

			pending := make(map[int64][]*pb.Memo)
			for _, memo := range memos {

				switch a := memo.Actual.(type) {
				// case *pb.Memo_SpawnEvent:
				// 	actual := a.SpawnEvent
				// 	mr.createMemos[actual.Nid] = memo
				case *pb.Memo_SpawnMissile:
					actual := a.SpawnMissile
					log.Println("SpawnMissile: ", actual)

					writeEvent(memo)

					mr.createMemos[actual.Nid] = memo
				case *pb.Memo_SpawnShip:
					actual := a.SpawnShip
					log.Println("SpawnShip: ", actual)

					writeEvent(memo)

					mr.createMemos[actual.Nid] = memo
				case *pb.Memo_DestroyEvent:
					actual := a.DestroyEvent
					log.Println("DestroyEvent", memo)

					writeEvent(memo)

					delete(mr.createMemos, actual.Nid)
				}

				for cid := range mr.outgoing {
					if isMemoRecipient(cid, memo) {
						pending[cid] = append(pending[cid], memo)
					}
				}
			}

			for cid, c := range mr.outgoing {
				combineToSend(c, pending[cid])
			}
			mr.outgoingLock.Unlock()
		}
	}()

	return mr
}

// TODO: Being lazy with client and server message passing: the clients, when
// sending a message to themselves (including broadcasts) should directly senda
// themselves the message.  So then the server here should take care to not
// send it back to that client (so it doesn't get the same message twice).
// Though also the server currently sends messages to itself through this router.
func (mr *memoRouter) connect(cid int64) (toSend chan []*pb.Memo, recieve func([]*pb.Memo)) {
	mr.outgoingLock.Lock()
	defer mr.outgoingLock.Unlock()

	if _, ok := mr.outgoing[cid]; ok {
		panic("Cid connected twice?")
	}

	toSend = make(chan []*pb.Memo, 1)
	mr.outgoing[cid] = toSend

	memos := []*pb.Memo{}
	for _, memo := range mr.createMemos {
		memos = append(memos, memo)
	}
	toSend <- memos

	recieve = func(memos []*pb.Memo) {
		combineToSend(mr.incoming, memos)
	}

	return toSend, recieve
}

func (mr *memoRouter) disconnect(cid int64) {
	// TODO: send disconnect memo
	mr.outgoingLock.Lock()
	defer mr.outgoingLock.Unlock()

	delete(mr.outgoing, cid)
}

func isMemoRecipient(cid int64, memo *pb.Memo) bool {
	switch r := memo.Recipient.(type) {
	case *pb.Memo_To:
		return cid == r.To
	case *pb.Memo_EveryoneBut:
		return cid != r.EveryoneBut
	case *pb.Memo_Everyone:
		return true
	}
	panic("Unknown recipient type")
}

///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////

func startAgones() (playerConnected func(), playerDisconnected func()) {
	waitForEmpty := &sync.WaitGroup{}

	{
		disabled, ok := os.LookupEnv("DISABLE_AGONES")
		if ok {
			if disabled == "true" {
				log.Println("Agones disabled")
				return func() {}, func() {}
			}
			log.Fatal("Unknown DISABLE_AGONES value:", disabled)
		}
	}

	log.Println("Starting Agones")
	a, err := agones.NewSDK()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		time.Sleep(3)
		a.Ready()
		for range time.Tick(time.Second) {
			a.Health()
		}
	}()

	var shutdown sync.Once
	var firstPlayerJoined sync.Once
	waitForFirstPlayer := make(chan struct{})
	a.WatchGameServer(func(gs *agonesSdk.GameServer) {
		if gs.GetStatus().GetState() == "Allocated" {
			shutdown.Do(func() {
				log.Println("Detected the server is allocated.")
				select {
				case <-time.After(time.Minute * 15):
					log.Println("Done waiting for first player to join.")
				case <-waitForFirstPlayer:
					log.Println("Detected first player joined")
				}
				log.Println("Waiting for all players to disconnect then shutting down.")
				waitForEmpty.Wait()
				log.Println("Server empty, shutting down.")
				a.Shutdown()
			})
		}
	})

	return func() {
			waitForEmpty.Add(1)
			firstPlayerJoined.Do(func() {
				close(waitForFirstPlayer)
			})
		}, func() {
			waitForEmpty.Done()
		}
}

func open(addr string) net.Conn{
	conn, err := net.Dial("tcp", listApiHost+":"+listApiPort)
	if err != nil {
		fmt.Println("Dialing "+addr+" failed:", err)
	}

	log.Println("Connected to tcp listener")

	return conn
}


func writeEvent(memo *pb.Memo) {

	c := open(listApiHost+":"+listApiPort)
	defer c.Close()

	// Marshall Protobuf to JSON
	marshaller := &jsonpb.Marshaler{}
	jsonStr, err := marshaller.MarshalToString(memo)
	if err != nil {
		log.Println("Can't convert to JSON: ", err)
	}
	log.Println("Sending to listener: ", jsonStr)
	_, err = c.Write([]byte(jsonStr))
	if err != nil {
		log.Println("Error writing data to tcp connection: ", err)
	}
}
