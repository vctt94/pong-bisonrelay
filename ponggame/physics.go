package ponggame

import (
	"math"
	"math/rand"

	"github.com/ndabAP/ping-pong/engine"
)

const (
	DEFAULT_FPS      = 60
	DEFAULT_VEL_INCR = 0.0005
	INPUT_BUF_SIZE   = 2 << 8

	baseline                 = 0
	default_padding          = 0
	canvas_border_correction = 1

	default_ball_x_vel_ratio = 0.25
	min_ball_y_vel_ratio     = 0.1
	max_y_vel_ratio          = 1.5
	initial_ball_y_vel       = 0.20

	magic_p = 3
)

// Helper function to check AABB intersection
func intersects(a, b Rect) bool {
	if math.Abs(a.Cx-b.Cx) > (a.HalfW + b.HalfW) {
		return false
	}
	if math.Abs(a.Cy-b.Cy) > (a.HalfH + b.HalfH) {
		return false
	}
	return true
}

// Bounding box helpers for game objects
func (e *CanvasEngine) ballRect() Rect {
	return Rect{
		Cx:    e.BallX + e.Game.Ball.Width*0.5,
		Cy:    e.BallY + e.Game.Ball.Height*0.5,
		HalfW: e.Game.Ball.Width * 0.5,
		HalfH: e.Game.Ball.Height * 0.5,
	}
}

func (e *CanvasEngine) p1Rect() Rect {
	return Rect{
		Cx:    e.P1X + e.Game.P1.Width*0.5,
		Cy:    e.P1Y + e.Game.P1.Height*0.5,
		HalfW: e.Game.P1.Width * 0.5,
		HalfH: e.Game.P1.Height * 0.5,
	}
}

func (e *CanvasEngine) p2Rect() Rect {
	return Rect{
		Cx:    e.P2X + e.Game.P2.Width*0.5,
		Cy:    e.P2Y + e.Game.P2.Height*0.5,
		HalfW: e.Game.P2.Width * 0.5,
		HalfH: e.Game.P2.Height * 0.5,
	}
}

// Wall rectangles
func (e *CanvasEngine) topRect() Rect {
	return Rect{
		Cx:    e.Game.Width * 0.5,
		Cy:    baseline + canvas_border_correction*0.5,
		HalfW: e.Game.Width * 0.5,
		HalfH: canvas_border_correction * 0.5,
	}
}

func (e *CanvasEngine) bottomRect() Rect {
	return Rect{
		Cx:    e.Game.Width * 0.5,
		Cy:    e.Game.Height - canvas_border_correction*0.5,
		HalfW: e.Game.Width * 0.5,
		HalfH: canvas_border_correction * 0.5,
	}
}

// tick calculates the next frame
func (e *CanvasEngine) tick() {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch e.detectColl() {

	case
		engine.CollP1Top,
		engine.CollP1Bottom,
		engine.CollP2Top,
		engine.CollP2Bottom:
		e.inverseBallXYVelocity().deOutOfBoundsBall()

	case
		engine.CollP1,
		engine.CollP2:
		e.inverseBallXVelocity().deOutOfBoundsBall()

	case
		engine.CollBottomLeft,
		engine.CollTopLeft:
		e.Err = engine.ErrP2Win
		return

	case
		engine.CollBottomRight,
		engine.CollTopRight:
		e.Err = engine.ErrP2Win
		return

	case
		engine.CollTop,
		engine.CollBottom:
		e.inverseBallYVelocity().deOutOfBoundsBall()

	case engine.CollLeft:
		e.Err = engine.ErrP2Win
		return

	case engine.CollRight:
		e.Err = engine.ErrP1Win
		return

	case engine.CollNone:
		fallthrough
	default:
	}

	e.advanceBall().deOutOfBoundsPlayers()
}

// State
func (e *CanvasEngine) ballDirP1() bool {
	return e.BallX <= e.Game.Width/2
}

func (e *CanvasEngine) ballDirP2() bool {
	return e.BallX >= e.Game.Width/2
}

// Collisions

// detectColl detects and returns a possible collision
func (e *CanvasEngine) detectColl() engine.Collision {
	br := e.ballRect()
	p1r := e.p1Rect()
	p2r := e.p2Rect()
	topr := e.topRect()
	bottomr := e.bottomRect()

	// Check paddle collisions first
	if intersects(br, p1r) {
		// Determine if it's a top/bottom collision by comparing centers
		if math.Abs(br.Cy-p1r.Cy) > p1r.HalfH*0.8 { // Using 0.8 as threshold
			if br.Cy < p1r.Cy {
				return engine.CollP1Top
			}
			return engine.CollP1Bottom
		}
		return engine.CollP1
	}

	if intersects(br, p2r) {
		// Similar top/bottom check for P2
		if math.Abs(br.Cy-p2r.Cy) > p2r.HalfH*0.8 {
			if br.Cy < p2r.Cy {
				return engine.CollP2Top
			}
			return engine.CollP2Bottom
		}
		return engine.CollP2
	}

	// Check wall collisions
	if intersects(br, topr) {
		if br.Cx <= p1r.Cx+p1r.HalfW {
			return engine.CollTopLeft
		}
		if br.Cx >= p2r.Cx-p2r.HalfW {
			return engine.CollTopRight
		}
		return engine.CollTop
	}

	if intersects(br, bottomr) {
		if br.Cx <= p1r.Cx+p1r.HalfW {
			return engine.CollBottomLeft
		}
		if br.Cx >= p2r.Cx-p2r.HalfW {
			return engine.CollBottomRight
		}
		return engine.CollBottom
	}

	// Check side walls (scoring)
	if br.Cx-br.HalfW <= 0 {
		return engine.CollLeft
	}
	if br.Cx+br.HalfW >= e.Game.Width {
		return engine.CollRight
	}

	return engine.CollNone
}

// Mutations

func (e *CanvasEngine) reset() *CanvasEngine {
	e.Err = nil
	return e.resetBall().resetPlayers()
}

func (e *CanvasEngine) resetBall() *CanvasEngine {
	// Center ball
	e.BallX = e.Game.Width / 2.0
	e.BallY = e.Game.Height / 2.0

	// Reset velocity multiplier to 1.0 at the start of each round
	e.VelocityMultiplier = 1.0

	// Random direction
	if rand.Intn(10) < 5 {
		e.BallXVelocity = -default_ball_x_vel_ratio * e.Game.Width
		y := min_ball_y_vel_ratio*e.Game.Height + rand.Float64()*((initial_ball_y_vel*e.Game.Height)-(min_ball_y_vel_ratio*e.Game.Height))
		e.BallYVelocity = -y
	} else {
		e.BallXVelocity = default_ball_x_vel_ratio * e.Game.Width
		y := min_ball_y_vel_ratio*e.Game.Height + rand.Float64()*((initial_ball_y_vel*e.Game.Height)-(min_ball_y_vel_ratio*e.Game.Height))
		e.BallYVelocity = y
	}
	return e
}

func (e *CanvasEngine) resetPlayers() *CanvasEngine {
	// P1
	e.P1X = 0 + default_padding
	e.P1Y = e.Game.Height/2 - e.Game.P1.Height/2
	e.P1YVelocity = 0

	// P2
	e.P2X = e.Game.Width - +e.Game.P1.Width - default_padding
	e.P2Y = e.Game.Height/2 - e.Game.P2.Height/2
	e.P2YVelocity = 0

	return e
}

// advanceBall advances the ball one tick or frame
func (e *CanvasEngine) advanceBall() *CanvasEngine {
	// Increase velocity multiplier gradually over time
	// Adjust the rate of increase (0.0001) to control how quickly the ball speeds up
	if e.VelocityIncrease > 0 {
		e.VelocityMultiplier += e.VelocityIncrease
	} else {
		e.VelocityMultiplier += DEFAULT_VEL_INCR
	}

	// Apply the velocity multiplier to the ball movement
	e.BallX += (e.BallXVelocity * e.VelocityMultiplier) / e.FPS
	e.BallY += (e.BallYVelocity * e.VelocityMultiplier) / e.FPS
	return e
}

func (e *CanvasEngine) p1Up() *CanvasEngine {
	e.P1YVelocity = max_y_vel_ratio * e.Game.Height
	e.P1Y -= e.P1YVelocity / e.FPS // Use velocity with timing, moving up (negative Y)
	return e
}

func (e *CanvasEngine) p1Down() *CanvasEngine {
	e.P1YVelocity = max_y_vel_ratio * e.Game.Height
	e.P1Y += e.P1YVelocity / e.FPS // Use velocity with timing, moving down (positive Y)
	return e
}

func (e *CanvasEngine) p2Up() *CanvasEngine {
	e.P2YVelocity = max_y_vel_ratio * e.Game.Height
	e.P2Y -= e.P2YVelocity / e.FPS // Use velocity with timing, moving up (negative Y)
	return e
}

func (e *CanvasEngine) p2Down() *CanvasEngine {
	e.P2YVelocity = max_y_vel_ratio * e.Game.Height
	e.P2Y += e.P2YVelocity / e.FPS // Use velocity with timing, moving down (positive Y)
	return e
}

func (e *CanvasEngine) inverseBallXYVelocity() *CanvasEngine {
	return e.inverseBallXVelocity().inverseBallYVelocity()
}

func (e *CanvasEngine) inverseBallXVelocity() *CanvasEngine {
	if e.BallXVelocity > 0 {
		e.BallXVelocity = e.BallXVelocity * -1
	} else {
		e.BallXVelocity = math.Abs(e.BallXVelocity)
	}
	return e
}

func (e *CanvasEngine) inverseBallYVelocity() *CanvasEngine {
	if e.BallYVelocity > 0 {
		e.BallYVelocity = e.BallYVelocity * -1
	} else {
		e.BallYVelocity = math.Abs(e.BallYVelocity)
	}
	return e
}

func (e *CanvasEngine) deOutOfBoundsPlayers() *CanvasEngine {
	// P1, top
	if e.P1Y-default_padding <= baseline {
		e.P1Y = baseline + default_padding
		e.P1YVelocity = 0
	}
	// P1, bottom
	if e.P1Y+e.Game.P1.Height >= e.Game.Height-default_padding {
		e.P1Y = e.Game.Height - e.Game.P1.Height - default_padding
		e.P1YVelocity = 0
	}
	// P2, top
	if e.P2Y-default_padding <= baseline {
		e.P2Y = baseline + default_padding
		e.P2YVelocity = 0
	}
	// P2, bottom
	if e.P2Y+e.Game.P2.Height >= e.Game.Height-default_padding {
		e.P2Y = e.Game.Height - e.Game.P2.Height - default_padding
		e.P2YVelocity = 0
	}
	return e
}

func (e *CanvasEngine) deOutOfBoundsBall() *CanvasEngine {
	// Top
	if e.BallY <= baseline {
		e.BallY = baseline - 1
	}
	// Bottom
	if e.BallY+e.Game.Ball.Height >= e.Game.Height {
		e.BallY = e.Game.Height - e.Game.Ball.Height - 1
	}
	// P1
	if e.BallX-e.Game.Ball.Width <= e.P1X {
		e.BallX = e.P1X + e.Game.P1.Width
	}
	// P2
	if e.BallX+e.Game.Ball.Width >= e.P2X {
		e.BallX = e.P2X - magic_p
	}
	return e
}
