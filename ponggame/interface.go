package ponggame

import (
	"context"
	"math"
	"sync"

	"github.com/companyzero/bisonrelay/client/clientintf"
	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

// Rect represents a bounding box with center position and half-dimensions
type Rect struct{ Cx, Cy, HalfW, HalfH float64 }

// -----------------------------------------------------------------------------
// Small math helpers
// -----------------------------------------------------------------------------

type Vec2 struct{ X, Y float64 }

func (a Vec2) Add(b Vec2) Vec2      { return Vec2{a.X + b.X, a.Y + b.Y} }
func (a Vec2) Sub(b Vec2) Vec2      { return Vec2{a.X - b.X, a.Y - b.Y} }
func (a Vec2) Scale(s float64) Vec2 { return Vec2{a.X * s, a.Y * s} }
func (a Vec2) Len() float64         { return math.Hypot(a.X, a.Y) }
func (a Vec2) Dot(b Vec2) float64   { return a.X*b.X + a.Y*b.Y }

// -----------------------------------------------------------------------------
// Data structures
// -----------------------------------------------------------------------------

type Vertex struct {
	Pos, Prev Vec2
	Vel       Vec2
	InvMass   float64 // 0 ⇒ pinned
	Radius    float64
}

type Constraint struct {
	I, J int     // vertex indices (J < 0 ⇒ plane)
	Rest float64 // soft‑spring rest length

	// plane description (for hard contact)
	N      Vec2    // outward normal
	Radius float64 // signed distance (n·x ≥ d)

	// augmented‑Lagrangian bookkeeping
	Hard   bool
	Lambda float64
	K      float64
	MinLam float64
	MaxLam float64
}

type AVBDPhysics struct {
	verts []Vertex
	cons  []Constraint

	gravity Vec2
	damping float64

	staticCount int // number of permanent (distance) constraints
}

type Player struct {
	ID *zkidentity.ShortID

	Nick           string
	BetAmt         int64
	PlayerNumber   int32 // 1 for player 1, 2 for player 2
	Score          int
	GameStream     pong.PongGame_StartGameStreamServer
	NotifierStream pong.PongGame_StartNtfnStreamServer
	Ready          bool

	// Per-player frame buffer to prevent one slow client from affecting others
	FrameCh chan []byte

	WR *WaitingRoom
}

func (p *Player) ResetPlayer() {
	p.GameStream = nil
	p.Score = 0
	p.PlayerNumber = 0
	p.BetAmt = 0
	p.Ready = false
	if p.FrameCh != nil {
		close(p.FrameCh)
		p.FrameCh = nil
	}
	p.WR = nil
}

type GameInstance struct {
	sync.RWMutex
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
	betAmt int64

	// Ready to play state
	PlayersReady     map[string]bool
	CountdownStarted bool
	CountdownValue   int
	GameReady        bool

	log slog.Logger
}

type WaitingRoom struct {
	sync.RWMutex
	Ctx          context.Context
	Cancel       context.CancelFunc
	ID           string
	HostID       *clientintf.UserID
	Players      []*Player
	BetAmount    int64
	ReservedTips []*types.ReceivedTip
}

type GameManager struct {
	sync.RWMutex

	ID             *zkidentity.ShortID
	Games          map[string]*GameInstance
	WaitingRooms   []*WaitingRoom
	PlayerSessions *PlayerSessions
	PlayerGameMap  map[zkidentity.ShortID]*GameInstance

	Log slog.Logger

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

	BallPos, BallVel Vec2
	// Replace individual position/velocity fields with vectors
	P1Pos, P2Pos Vec2
	P1Vel, P2Vel Vec2

	// Error of the current tick
	Err error

	// Engine debug state
	log slog.Logger

	mu sync.RWMutex

	phy *AVBDPhysics
}

// StartGameStreamRequest encapsulates the data needed to start a game stream.
type StartGameStreamRequest struct {
	ClientID zkidentity.ShortID
	Stream   pong.PongGame_StartGameStreamServer
	MinBet   float64
	IsF2P    bool
	Log      slog.Logger
}
