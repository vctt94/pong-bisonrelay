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

	canvas_border_correction = 1

	default_ball_x_vel_ratio = 0.25
	min_ball_y_vel_ratio     = 0.1
	y_vel_ratio              = 2
	initial_ball_y_vel       = 0.20

	magic_p = 1
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
		Cx:    e.BallPos.X + e.Game.Ball.Width*0.5,
		Cy:    e.BallPos.Y + e.Game.Ball.Height*0.5,
		HalfW: e.Game.Ball.Width * 0.5,
		HalfH: e.Game.Ball.Height * 0.5,
	}
}

func (e *CanvasEngine) p1Rect() Rect {
	return Rect{
		Cx:    e.P1Pos.X + e.Game.P1.Width*0.5,
		Cy:    e.P1Pos.Y + e.Game.P1.Height*0.5,
		HalfW: e.Game.P1.Width * 0.5,
		HalfH: e.Game.P1.Height * 0.5,
	}
}

func (e *CanvasEngine) p2Rect() Rect {
	return Rect{
		Cx:    e.P2Pos.X + e.Game.P2.Width*0.5,
		Cy:    e.P2Pos.Y + e.Game.P2.Height*0.5,
		HalfW: e.Game.P2.Width * 0.5,
		HalfH: e.Game.P2.Height * 0.5,
	}
}

// Wall rectangles
func (e *CanvasEngine) topRect() Rect {
	return Rect{
		Cx:    e.Game.Width * 0.5,
		Cy:    canvas_border_correction * 0.5,
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
	// Only lock when actually updating shared state
	e.mu.RLock()
	collision := e.detectColl()
	e.mu.RUnlock()

	// Process collision result
	switch collision {
	case engine.CollP1Top,
		engine.CollP1Bottom,
		engine.CollP2Top,
		engine.CollP2Bottom:
		e.mu.Lock()
		e.handlePaddleEdgeHit().deOutOfBoundsBall()
		e.mu.Unlock()
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

	// Update ball position
	e.mu.Lock()
	e.advanceBall().deOutOfBoundsPlayers()
	e.mu.Unlock()
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
	e.BallPos = Vec2{
		X: e.Game.Width / 2.0,
		Y: e.Game.Height / 2.0,
	}

	// Reset velocity multiplier to 1.0 at the start of each round
	e.VelocityMultiplier = 1.0

	// Random direction
	xVel := default_ball_x_vel_ratio * e.Game.Width
	yVel := min_ball_y_vel_ratio*e.Game.Height +
		rand.Float64()*((initial_ball_y_vel*e.Game.Height)-(min_ball_y_vel_ratio*e.Game.Height))

	if rand.Intn(10) < 5 {
		e.BallVel = Vec2{-xVel, -yVel}
	} else {
		e.BallVel = Vec2{xVel, yVel}
	}
	return e
}

func (e *CanvasEngine) resetPlayers() *CanvasEngine {
	// P1
	e.P1Pos = Vec2{
		X: 0,
		Y: e.Game.Height/2 - e.Game.P1.Height/2,
	}
	e.P1Vel = Vec2{0, 0}

	// P2
	e.P2Pos = Vec2{
		X: e.Game.Width - e.Game.P1.Width,
		Y: e.Game.Height/2 - e.Game.P2.Height/2,
	}
	e.P2Vel = Vec2{0, 0}

	return e
}

// advanceBall advances the ball one tick or frame
func (e *CanvasEngine) advanceBall() *CanvasEngine {
	// Increase velocity multiplier gradually over time
	if e.VelocityIncrease > 0 {
		e.VelocityMultiplier += e.VelocityIncrease
	} else {
		e.VelocityMultiplier += DEFAULT_VEL_INCR
	}

	// Apply the velocity multiplier to the ball movement
	dt := 1.0 / e.FPS
	velocityThisTick := e.BallVel.Scale(e.VelocityMultiplier * dt)
	e.BallPos = e.BallPos.Add(velocityThisTick)
	return e
}

func (e *CanvasEngine) p1Up() *CanvasEngine {
	speed := y_vel_ratio * e.Game.Height
	e.P1Vel = Vec2{0, -speed}
	dt := 1.0 / e.FPS
	e.P1Pos = e.P1Pos.Add(e.P1Vel.Scale(dt))
	return e
}

func (e *CanvasEngine) p1Down() *CanvasEngine {
	speed := y_vel_ratio * e.Game.Height
	e.P1Vel = Vec2{0, speed}
	dt := 1.0 / e.FPS
	e.P1Pos = e.P1Pos.Add(e.P1Vel.Scale(dt))
	return e
}

func (e *CanvasEngine) p2Up() *CanvasEngine {
	speed := y_vel_ratio * e.Game.Height
	e.P2Vel = Vec2{0, -speed}
	dt := 1.0 / e.FPS
	e.P2Pos = e.P2Pos.Add(e.P2Vel.Scale(dt))
	return e
}

func (e *CanvasEngine) p2Down() *CanvasEngine {
	speed := y_vel_ratio * e.Game.Height
	e.P2Vel = Vec2{0, speed}
	dt := 1.0 / e.FPS
	e.P2Pos = e.P2Pos.Add(e.P2Vel.Scale(dt))
	return e
}

func (e *CanvasEngine) inverseBallXYVelocity() *CanvasEngine {
	return e.inverseBallXVelocity().inverseBallYVelocity()
}

func (e *CanvasEngine) inverseBallXVelocity() *CanvasEngine {
	e.BallVel.X *= -1
	return e
}

func (e *CanvasEngine) inverseBallYVelocity() *CanvasEngine {
	e.BallVel.Y *= -1
	return e
}

func (e *CanvasEngine) deOutOfBoundsPlayers() *CanvasEngine {
	// P1, top
	if e.P1Pos.Y <= 0 {
		e.P1Pos.Y = 0
		e.P1Vel.Y = 0
	}
	// P1, bottom
	if e.P1Pos.Y+e.Game.P1.Height >= e.Game.Height {
		e.P1Pos.Y = e.Game.Height - e.Game.P1.Height
		e.P1Vel.Y = 0
	}
	// P2, top
	if e.P2Pos.Y <= 0 {
		e.P2Pos.Y = 0
		e.P2Vel.Y = 0
	}
	// P2, bottom
	if e.P2Pos.Y+e.Game.P2.Height >= e.Game.Height-0 {
		e.P2Pos.Y = e.Game.Height - e.Game.P2.Height - 0
		e.P2Vel.Y = 0
	}
	return e
}

func (e *CanvasEngine) deOutOfBoundsBall() *CanvasEngine {
	// Top
	if e.BallPos.Y <= 0 {
		e.BallPos.Y = -1
	}
	// Bottom
	if e.BallPos.Y+e.Game.Ball.Height >= e.Game.Height {
		e.BallPos.Y = e.Game.Height - e.Game.Ball.Height - 1
	}
	// P1
	if e.BallPos.X-e.Game.Ball.Width <= e.P1Pos.X {
		e.BallPos.X = e.P1Pos.X + e.Game.P1.Width
	}
	// P2
	if e.BallPos.X+e.Game.Ball.Width >= e.P2Pos.X {
		e.BallPos.X = e.P2Pos.X - magic_p
	}
	return e
}

func (e *CanvasEngine) handlePaddleEdgeHit() *CanvasEngine {
	// First, invert the X direction as we always want the ball to bounce back
	e.BallVel.X *= -1

	// Calculate current ball speed (magnitude of velocity)
	currentSpeed := math.Sqrt(e.BallVel.X*e.BallVel.X + e.BallVel.Y*e.BallVel.Y)

	// For edge hits, we want a steeper angle but maintain similar speed
	// Use a 60-degree angle (approximately 0.866 for x and 0.5 for y components)
	normalizedX := math.Abs(e.BallVel.X) / currentSpeed

	// Determine the direction of Y velocity based on which edge was hit
	yDirection := 1.0
	if e.BallVel.Y < 0 {
		yDirection = -1.0
	}

	// Set new velocities while maintaining approximate original speed
	e.BallVel.X = normalizedX * currentSpeed * math.Copysign(1, e.BallVel.X)
	e.BallVel.Y = 0.5 * currentSpeed * yDirection // Use 0.5 for a consistent but not too extreme angle

	return e
}
