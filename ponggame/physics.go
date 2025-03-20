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
	switch {
	case e.isCollBottomLeft():
		return engine.CollBottomLeft

	case e.isCollTopLeft():
		return engine.CollTopLeft

	case e.isCollBottomRight():
		return engine.CollBottomRight

	case e.isCollTopRight():
		return engine.CollTopRight

	case e.isCollP1Bottom():
		return engine.CollP1Bottom

	case e.isCollP1Top():
		return engine.CollP1Top

	case e.isCollP2Bottom():
		return engine.CollP2Bottom

	case e.isCollP2Top():
		return engine.CollP2Top

	case e.isCollP1():
		return engine.CollP1

	case e.isCollP2():
		return engine.CollP2

	case e.isCollBottom():
		return engine.CollBottom

	case e.isCollTop():
		return engine.CollTop

	case e.isCollLeft():
		return engine.CollLeft

	case e.isCollRight():
		return engine.CollRight

	default:
		return engine.CollNone
	}
}

func (e *CanvasEngine) isCollP1() bool {
	x := e.BallX <= (e.P1X + e.Game.P1.Width + 1)

	y1 := e.BallY <= (e.P1Y + e.Game.P1.Height)
	y2 := (e.BallY + e.Game.Ball.Height) >= e.P1Y

	y := y1 && y2
	return x && y
}

func (e *CanvasEngine) isCollP2() bool {
	x := (e.BallX + e.Game.Ball.Width) >= e.P2X

	y1 := e.BallY <= (e.P2Y + e.Game.P2.Height)
	y2 := (e.BallY + e.Game.Ball.Height) >= e.P2Y

	y := y1 && y2
	return x && y
}

func (e *CanvasEngine) isCollTop() bool {
	return e.BallY <= baseline+e.Game.Ball.Height+canvas_border_correction
}

func (e *CanvasEngine) isCollBottom() bool {
	return e.BallY+e.Game.Ball.Height >= e.Game.Height-canvas_border_correction
}

func (e *CanvasEngine) isCollLeft() bool {
	return e.BallX-e.Game.Ball.Height-canvas_border_correction <= 0
}

func (e *CanvasEngine) isCollRight() bool {
	return e.BallX+e.Game.Ball.Height+canvas_border_correction >= e.Game.Width
}

func (e *CanvasEngine) isCollP1Top() bool {
	return e.isCollP1() && e.isCollTop()
}

func (e *CanvasEngine) isCollP2Top() bool {
	return e.isCollP2() && e.isCollTop()
}

func (e *CanvasEngine) isCollP1Bottom() bool {
	return e.isCollP1() && e.isCollBottom()
}

func (e *CanvasEngine) isCollP2Bottom() bool {
	return e.isCollP2() && e.isCollBottom()
}

func (e *CanvasEngine) isCollTopLeft() bool {
	return e.isCollTop() && e.isCollLeft()
}

func (e *CanvasEngine) isCollBottomLeft() bool {
	return e.isCollBottom() && e.isCollLeft()
}

func (e *CanvasEngine) isCollTopRight() bool {
	return e.isCollTop() && e.isCollRight()
}

func (e *CanvasEngine) isCollBottomRight() bool {
	return e.isCollBottom() && e.isCollRight()
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
	e.mu.Lock()
	defer e.mu.Unlock()
	e.P1YVelocity = max_y_vel_ratio * e.Game.Height
	e.P1Y -= e.P1YVelocity / e.FPS // Use velocity with timing, moving up (negative Y)
	return e
}

func (e *CanvasEngine) p1Down() *CanvasEngine {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.P1YVelocity = max_y_vel_ratio * e.Game.Height
	e.P1Y += e.P1YVelocity / e.FPS // Use velocity with timing, moving down (positive Y)
	return e
}

func (e *CanvasEngine) p2Up() *CanvasEngine {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.P2YVelocity = max_y_vel_ratio * e.Game.Height
	e.P2Y -= e.P2YVelocity / e.FPS // Use velocity with timing, moving up (negative Y)
	return e
}

func (e *CanvasEngine) p2Down() *CanvasEngine {
	e.mu.Lock()
	defer e.mu.Unlock()
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
	if e.BallX-e.Game.Ball.Width <= e.P1X+e.Game.P1.Width {
		e.BallX = e.P1X + e.Game.P1.Width + magic_p
	}
	// P2
	if e.BallX+e.Game.Ball.Width >= e.P2X {
		e.BallX = e.P2X - magic_p
	}
	return e
}
