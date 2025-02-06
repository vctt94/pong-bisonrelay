package ponggame

import (
	"context"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

type Player struct {
	ID *zkidentity.ShortID

	Nick           string
	BetAmt         float64
	PlayerNumber   int32 // 1 for player 1, 2 for player 2
	Score          int
	GameStream     pong.PongGame_StartGameStreamServer
	NotifierStream pong.PongGame_StartNtfnStreamServer
	Ready          bool
}

func (p *Player) ResetPlayer() {
	p.Score = 0
	p.PlayerNumber = 0
	p.BetAmt = 0
	p.Ready = false
}

type GameInstance struct {
	sync.Mutex
	Id          string
	engine      *CanvasEngine
	Framesch    chan []byte
	Inputch     chan []byte
	roundResult chan int32
	Players     []*Player
	cleanedUp   bool
	Running     bool
	ctx         context.Context
	cancel      context.CancelFunc
	Winner      *zkidentity.ShortID
	// betAmt sum of total bets
	betAmt float64

	log slog.Logger
}

type WaitingRoom struct {
	sync.RWMutex
	Ctx       context.Context
	Cancel    context.CancelFunc
	ID        string
	HostID    *clientintf.UserID
	Players   []*Player
	BetAmount float64
}

type GameManager struct {
	sync.RWMutex

	ID             *zkidentity.ShortID
	Games          map[string]*GameInstance
	WaitingRooms   []*WaitingRoom
	PlayerSessions *PlayerSessions

	Debug slog.Level
	Log   slog.Logger

	// Callback for waiting room removal notifications
	OnWaitingRoomRemoved func(*pong.WaitingRoom)
}

// CanvasEngine is a ping-pong engine for browsers with Canvas support
type CanvasEngine struct {
	// Static
	FPS, TPS float64

	Game engine.Game

	// State
	P1Score, P2Score int

	BallX, BallY       float64
	P1X, P1Y, P2X, P2Y float64

	P1YVelocity, P2YVelocity     float64
	BallXVelocity, BallYVelocity float64

	// Error of the current tick
	Err error

	// Engine debug state
	log slog.Logger
}

// StartGameStreamRequest encapsulates the data needed to start a game stream.
type StartGameStreamRequest struct {
	ClientID zkidentity.ShortID
	Stream   pong.PongGame_StartGameStreamServer
	MinBet   float64
	IsF2P    bool
	Log      slog.Logger
}
