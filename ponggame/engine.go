package ponggame

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/decred/slog"
	"golang.org/x/exp/rand"
	"google.golang.org/protobuf/proto"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"

	"github.com/ndabAP/ping-pong/engine"
)

// Pool for GameUpdate objects to reduce allocations
var gameUpdatePool = sync.Pool{
	New: func() interface{} {
		return &pong.GameUpdate{}
	},
}

// New returns a new Canvas engine pre‑wired with the AVBD physics solver.
func New(g engine.Game) *CanvasEngine {
	e := new(CanvasEngine)
	e.Game = g
	e.FPS = DEFAULT_FPS
	e.TPS = 1000.0 / e.FPS

	// zero everything explicitly
	e.P1Vel = Vec2{}
	e.P2Vel = Vec2{}
	e.BallVel = Vec2{}

	// place objects at their default positions
	e.reset()

	// build physics *after* reset so it sees the correct pose
	e.phy = NewAVBDPhysics(e)
	return e
}

// -----------------------------------------------------------------------------
//  Public configuration helpers
// -----------------------------------------------------------------------------

func (e *CanvasEngine) SetLogger(log slog.Logger) *CanvasEngine {
	e.log = log
	return e
}

func (e *CanvasEngine) SetFPS(fps uint) *CanvasEngine {
	if fps == 0 {
		panic("fps must be greater than zero")
	}
	e.log.Debugf("fps %d", fps)
	e.FPS = float64(fps)
	e.TPS = 1000.0 / e.FPS
	return e
}

// Error returns the Canvas engine's last error (round finished).
func (e *CanvasEngine) Error() error { return e.Err }

// -----------------------------------------------------------------------------
//
//	Round loop
//
// -----------------------------------------------------------------------------
func (e *CanvasEngine) NewRound(ctx context.Context, framesch chan<- []byte, inputch <-chan []byte, roundResult chan<- int32) {
	time.Sleep(time.Second)
	e.reset()

	// frame pump
	go func() {
		tick := time.NewTicker(time.Duration(1000.0/e.FPS) * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				e.tick()
				if errors.Is(e.Err, engine.ErrP1Win) {
					e.P1Score++
					roundResult <- 1
					return
				}
				if errors.Is(e.Err, engine.ErrP2Win) {
					e.P2Score++
					roundResult <- 2
					return
				}
				gu := gameUpdatePool.Get().(*pong.GameUpdate)
				gu.Reset()
				gu.GameWidth, gu.GameHeight = e.Game.Width, e.Game.Height
				gu.P1Width, gu.P1Height = e.Game.P1.Width, e.Game.P1.Height
				gu.P2Width, gu.P2Height = e.Game.P2.Width, e.Game.P2.Height
				gu.BallWidth, gu.BallHeight = e.Game.Ball.Width, e.Game.Ball.Height
				gu.P1Score, gu.P2Score = int32(e.P1Score), int32(e.P2Score)
				gu.BallX, gu.BallY = e.BallPos.X, e.BallPos.Y
				gu.P1X, gu.P1Y = e.P1Pos.X, e.P1Pos.Y
				gu.P2X, gu.P2Y = e.P2Pos.X, e.P2Pos.Y
				gu.P1YVelocity, gu.P2YVelocity = e.P1Vel.Y, e.P2Vel.Y
				gu.BallXVelocity, gu.BallYVelocity = e.BallVel.X, e.BallVel.Y
				gu.Fps, gu.Tps = e.FPS, e.TPS
				data, _ := proto.Marshal(gu)
				gameUpdatePool.Put(gu)
				select {
				case framesch <- data:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// input pump
	go func() {
		for {
			select {
			case raw, ok := <-inputch:
				if !ok || len(raw) == 0 {
					return
				}
				in := &pong.PlayerInput{}
				if err := proto.Unmarshal(raw, in); err != nil {
					continue
				}
				if in.PlayerNumber == 1 {
					e.handleP1Input(in.Input)
				} else {
					e.handleP2Input(in.Input)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// -------------------- reset helpers -----------------------------------------

func (e *CanvasEngine) reset() *CanvasEngine {
	e.Err = nil
	e.resetBall()
	e.resetPlayers()
	e.phy = NewAVBDPhysics(e)
	return e
}

func (e *CanvasEngine) resetBall() {
	cx, cy := e.Game.Width*0.5, e.Game.Height*0.5
	e.BallPos = Vec2{cx - e.Game.Ball.Width*0.5, cy - e.Game.Ball.Height*0.5}
	vx := initial_ball_x_vel * e.Game.Width
	vy := initial_ball_y_vel * e.Game.Height
	if rand.Intn(2) == 0 {
		vx = -vx
	}
	if rand.Intn(2) == 0 {
		vy = -vy
	}
	e.BallVel = Vec2{vx, vy}
}

func (e *CanvasEngine) resetPlayers() {
	h := e.Game.Height * 0.5
	p1x := e.Game.P1.Width * 1.5
	p2x := e.Game.Width - e.Game.P2.Width*1.5
	e.P1Pos = Vec2{p1x - e.Game.P1.Width*0.5, h - e.Game.P1.Height*0.5}
	e.P2Pos = Vec2{p2x - e.Game.P2.Width*0.5, h - e.Game.P2.Height*0.5}
	e.P1Vel, e.P2Vel = Vec2{}, Vec2{}
}

// -------------------- input helpers ----------------------------------------

func (e *CanvasEngine) handleP1Input(k string) {
	switch k {
	case "ArrowUp":
		e.p1Up()
	case "ArrowDown":
		e.p1Down()
	case "ArrowUpStop":
		if e.P1Vel.Y < 0 {
			e.P1Vel.Y = 0
		}
	case "ArrowDownStop":
		if e.P1Vel.Y > 0 {
			e.P1Vel.Y = 0
		}
	}
}
func (e *CanvasEngine) handleP2Input(k string) {
	switch k {
	case "ArrowUp":
		e.p2Up()
	case "ArrowDown":
		e.p2Down()
	case "ArrowUpStop":
		if e.P2Vel.Y < 0 {
			e.P2Vel.Y = 0
		}
	case "ArrowDownStop":
		if e.P2Vel.Y > 0 {
			e.P2Vel.Y = 0
		}
	}
}

// -------------------- paddle motion ----------------------------------------

func (e *CanvasEngine) p1Up()   { e.P1Vel = Vec2{0, -y_vel_ratio * e.Game.Height} }
func (e *CanvasEngine) p1Down() { e.P1Vel = Vec2{0, y_vel_ratio * e.Game.Height} }
func (e *CanvasEngine) p2Up()   { e.P2Vel = Vec2{0, -y_vel_ratio * e.Game.Height} }
func (e *CanvasEngine) p2Down() { e.P2Vel = Vec2{0, y_vel_ratio * e.Game.Height} }

// -------------------- collision helpers (reuse physics code) ---------------

func (e *CanvasEngine) ballRect() Rect {
	return Rect{e.BallPos.X + e.Game.Ball.Width*0.5, e.BallPos.Y + e.Game.Ball.Height*0.5, e.Game.Ball.Width * 0.5, e.Game.Ball.Height * 0.5}
}
func (e *CanvasEngine) p1Rect() Rect {
	return Rect{e.P1Pos.X + e.Game.P1.Width*0.5, e.P1Pos.Y + e.Game.P1.Height*0.5, e.Game.P1.Width * 0.5, e.Game.P1.Height * 0.5}
}
func (e *CanvasEngine) p2Rect() Rect {
	return Rect{e.P2Pos.X + e.Game.P2.Width*0.5, e.P2Pos.Y + e.Game.P2.Height*0.5, e.Game.P2.Width * 0.5, e.Game.P2.Height * 0.5}
}
func (e *CanvasEngine) topRect() Rect {
	return Rect{e.Game.Width * 0.5, canvas_border_correction * 0.5, e.Game.Width * 0.5, canvas_border_correction * 0.5}
}
func (e *CanvasEngine) bottomRect() Rect {
	return Rect{e.Game.Width * 0.5, e.Game.Height - canvas_border_correction*0.5, e.Game.Width * 0.5, canvas_border_correction * 0.5}
}

func intersects(a, b Rect) bool {
	return math.Abs(a.Cx-b.Cx) <= a.HalfW+b.HalfW && math.Abs(a.Cy-b.Cy) <= a.HalfH+b.HalfH
}

func (e *CanvasEngine) detectColl() engine.Collision {
	br, p1r, p2r := e.ballRect(), e.p1Rect(), e.p2Rect()
	tr, brw := e.topRect(), e.bottomRect()

	if intersects(br, p1r) {
		if math.Abs(br.Cy-p1r.Cy) > p1r.HalfH*0.8 {
			if br.Cy < p1r.Cy {
				return engine.CollP1Top
			}
			return engine.CollP1Bottom
		}
		return engine.CollP1
	}
	if intersects(br, p2r) {
		if math.Abs(br.Cy-p2r.Cy) > p2r.HalfH*0.8 {
			if br.Cy < p2r.Cy {
				return engine.CollP2Top
			}
			return engine.CollP2Bottom
		}
		return engine.CollP2
	}
	if intersects(br, tr) {
		if br.Cx <= p1r.Cx+p1r.HalfW {
			return engine.CollTopLeft
		}
		if br.Cx >= p2r.Cx-p2r.HalfW {
			return engine.CollTopRight
		}
		return engine.CollTop
	}
	if intersects(br, brw) {
		if br.Cx <= p1r.Cx+p1r.HalfW {
			return engine.CollBottomLeft
		}
		if br.Cx >= p2r.Cx-p2r.HalfW {
			return engine.CollBottomRight
		}
		return engine.CollBottom
	}
	if br.Cx-br.HalfW <= 0 {
		return engine.CollLeft
	}
	if br.Cx+br.HalfW >= e.Game.Width {
		return engine.CollRight
	}
	return engine.CollNone
}

func (e *CanvasEngine) deOutOfBoundsPlayers() {
	p1r, p2r := e.p1Rect(), e.p2Rect()
	if p1r.Cy-p1r.HalfH <= 0 {
		p1r.Cy = p1r.HalfH
		e.P1Pos.Y = p1r.Cy - p1r.HalfH
		e.P1Vel.Y = 0
	}
	if p1r.Cy+p1r.HalfH >= e.Game.Height {
		p1r.Cy = e.Game.Height - p1r.HalfH
		e.P1Pos.Y = p1r.Cy - p1r.HalfH
		e.P1Vel.Y = 0
	}
	if p2r.Cy-p2r.HalfH <= 0 {
		p2r.Cy = p2r.HalfH
		e.P2Pos.Y = p2r.Cy - p2r.HalfH
		e.P2Vel.Y = 0
	}
	if p2r.Cy+p2r.HalfH >= e.Game.Height {
		p2r.Cy = e.Game.Height - p2r.HalfH
		e.P2Pos.Y = p2r.Cy - p2r.HalfH
		e.P2Vel.Y = 0
	}
}

func (e *CanvasEngine) deOutOfBoundsBall() {
	br := e.ballRect()
	if br.Cy-br.HalfH <= 0 {
		br.Cy = br.HalfH
		e.BallPos.Y = br.Cy - br.HalfH
	}
	if br.Cy+br.HalfH >= e.Game.Height {
		br.Cy = e.Game.Height - br.HalfH
		e.BallPos.Y = br.Cy - br.HalfH
	}
	p1r, p2r := e.p1Rect(), e.p2Rect()
	if br.Cx-br.HalfW <= p1r.Cx+p1r.HalfW {
		dx := (br.HalfW + p1r.HalfW) - math.Abs(br.Cx-p1r.Cx)
		if dx > 0 {
			if br.Cx > p1r.Cx {
				br.Cx += dx
			} else {
				br.Cx = p1r.Cx + p1r.HalfW + br.HalfW
			}
			e.BallPos.X = br.Cx - br.HalfW
		}
	}
	if br.Cx+br.HalfW >= p2r.Cx-p2r.HalfW {
		dx := (br.HalfW + p2r.HalfW) - math.Abs(br.Cx-p2r.Cx)
		if dx > 0 {
			if br.Cx < p2r.Cx {
				br.Cx -= dx
			} else {
				br.Cx = p2r.Cx - p2r.HalfW - br.HalfW
			}
			e.BallPos.X = br.Cx - br.HalfW
		}
	}
}

// -------------------- public state snapshot (unchanged) --------------------

func (e *CanvasEngine) State() struct{ PaddleWidth, PaddleHeight, BallWidth, BallHeight, P1PosX, P1PosY, P2PosX, P2PosY, BallPosX, BallPosY, BallVelX, BallVelY, FPS, TPS float64 } {
	return struct{ PaddleWidth, PaddleHeight, BallWidth, BallHeight, P1PosX, P1PosY, P2PosX, P2PosY, BallPosX, BallPosY, BallVelX, BallVelY, FPS, TPS float64 }{
		PaddleWidth: e.Game.P1.Width, PaddleHeight: e.Game.P1.Height, BallWidth: e.Game.Ball.Width, BallHeight: e.Game.Ball.Height,
		P1PosX: e.P1Pos.X, P1PosY: e.P1Pos.Y, P2PosX: e.P2Pos.X, P2PosY: e.P2Pos.Y, BallPosX: e.BallPos.X, BallPosY: e.BallPos.Y,
		BallVelX: e.BallVel.X, BallVelY: e.BallVel.Y, FPS: e.FPS, TPS: e.TPS}
}

// -----------------------------------------------------------------------------
//  Tick – one simulation / gameplay step
// -----------------------------------------------------------------------------

func (e *CanvasEngine) tick() {
	dt := 1.0 / e.FPS

	// 1. integrar os paddles (cinemática simples)
	e.P1Pos = e.P1Pos.Add(e.P1Vel.Scale(dt))
	e.P2Pos = e.P2Pos.Add(e.P2Vel.Scale(dt))

	// 2. step de física AVBD
	StepPhysics(e, dt)

	// 3. detectar colisão depois das posições atualizadas
	coll := e.detectColl()

	// 4. reagir à colisão (bounce, ponto, etc.)
	switch coll {
	case engine.CollP1, engine.CollP1Top, engine.CollP1Bottom,
		engine.CollP2, engine.CollP2Top, engine.CollP2Bottom:
		e.applyPaddleBounce(coll) // calcula novo vetor
		e.BallVel = clampBallSpeed(e.BallVel, e.Game.Width)

	case engine.CollTop:
		e.BallVel.Y = -math.Abs(e.BallVel.Y)
		bounceWall(e, +1) // teto → Vy positivo depois de inverter

	case engine.CollBottom:
		e.BallVel.Y = math.Abs(e.BallVel.Y)
		bounceWall(e, -1) // piso → Vy negativo depois de inverter

	case engine.CollTopLeft, engine.CollBottomLeft, engine.CollLeft:
		e.Err = engine.ErrP2Win
		return
	case engine.CollTopRight, engine.CollBottomRight, engine.CollRight:
		e.Err = engine.ErrP1Win
		return
	}

	// 5. correções de segurança de posição
	e.deOutOfBoundsBall()
	e.deOutOfBoundsPlayers()

	// 6. clamp global (última barreira de segurança)
	e.BallVel = clampBallSpeed(e.BallVel, e.Game.Width)
}
