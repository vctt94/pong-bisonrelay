package ponggame

import (
	"testing"

	"github.com/decred/slog"
	"github.com/ndabAP/ping-pong/engine"
	"github.com/stretchr/testify/assert"
)

func createTestEngine() *CanvasEngine {
	game := engine.NewGame(
		800, 400, // width, height
		engine.NewPlayer(10, 80), // P1: width, height
		engine.NewPlayer(10, 80), // P2: width, height
		engine.NewBall(10, 10),   // Ball: width, height
	)

	e := New(game)
	log := slog.Disabled
	e.SetLogger(log)
	e.reset()
	return e
}

func TestIntersects(t *testing.T) {
	tests := []struct {
		name string
		a    Rect
		b    Rect
		want bool
	}{
		{
			name: "overlapping rectangles",
			a:    Rect{Cx: 100, Cy: 100, HalfW: 50, HalfH: 50},
			b:    Rect{Cx: 120, Cy: 120, HalfW: 50, HalfH: 50},
			want: true,
		},
		{
			name: "non-overlapping rectangles",
			a:    Rect{Cx: 100, Cy: 100, HalfW: 20, HalfH: 20},
			b:    Rect{Cx: 200, Cy: 200, HalfW: 20, HalfH: 20},
			want: false,
		},
		{
			name: "touching rectangles",
			a:    Rect{Cx: 100, Cy: 100, HalfW: 25, HalfH: 25},
			b:    Rect{Cx: 150, Cy: 100, HalfW: 25, HalfH: 25},
			want: true,
		},
		{
			name: "identical rectangles",
			a:    Rect{Cx: 100, Cy: 100, HalfW: 50, HalfH: 50},
			b:    Rect{Cx: 100, Cy: 100, HalfW: 50, HalfH: 50},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intersects(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCanvasEngine_BallRect(t *testing.T) {
	e := createTestEngine()
	e.BallPos = Vec2{X: 100, Y: 50}

	rect := e.ballRect()

	// Ball position + half width/height should give center
	expectedCx := e.BallPos.X + e.Game.Ball.Width*0.5
	expectedCy := e.BallPos.Y + e.Game.Ball.Height*0.5

	assert.Equal(t, expectedCx, rect.Cx)
	assert.Equal(t, expectedCy, rect.Cy)
	assert.Equal(t, e.Game.Ball.Width*0.5, rect.HalfW)
	assert.Equal(t, e.Game.Ball.Height*0.5, rect.HalfH)
}

func TestCanvasEngine_PaddleRects(t *testing.T) {
	e := createTestEngine()
	e.P1Pos = Vec2{X: 10, Y: 100}
	e.P2Pos = Vec2{X: 780, Y: 200}

	// Test P1 rect
	p1Rect := e.p1Rect()
	expectedP1Cx := e.P1Pos.X + e.Game.P1.Width*0.5
	expectedP1Cy := e.P1Pos.Y + e.Game.P1.Height*0.5

	assert.Equal(t, expectedP1Cx, p1Rect.Cx)
	assert.Equal(t, expectedP1Cy, p1Rect.Cy)
	assert.Equal(t, e.Game.P1.Width*0.5, p1Rect.HalfW)
	assert.Equal(t, e.Game.P1.Height*0.5, p1Rect.HalfH)

	// Test P2 rect
	p2Rect := e.p2Rect()
	expectedP2Cx := e.P2Pos.X + e.Game.P2.Width*0.5
	expectedP2Cy := e.P2Pos.Y + e.Game.P2.Height*0.5

	assert.Equal(t, expectedP2Cx, p2Rect.Cx)
	assert.Equal(t, expectedP2Cy, p2Rect.Cy)
	assert.Equal(t, e.Game.P2.Width*0.5, p2Rect.HalfW)
	assert.Equal(t, e.Game.P2.Height*0.5, p2Rect.HalfH)
}

func TestCanvasEngine_WallRects(t *testing.T) {
	e := createTestEngine()

	// Test top wall
	topRect := e.topRect()
	assert.Equal(t, e.Game.Width*0.5, topRect.Cx)
	assert.Equal(t, canvas_border_correction*0.5, topRect.Cy)
	assert.Equal(t, e.Game.Width*0.5, topRect.HalfW)
	assert.Equal(t, canvas_border_correction*0.5, topRect.HalfH)

	// Test bottom wall
	bottomRect := e.bottomRect()
	assert.Equal(t, e.Game.Width*0.5, bottomRect.Cx)
	assert.Equal(t, e.Game.Height-canvas_border_correction*0.5, bottomRect.Cy)
	assert.Equal(t, e.Game.Width*0.5, bottomRect.HalfW)
	assert.Equal(t, canvas_border_correction*0.5, bottomRect.HalfH)
}

func TestCanvasEngine_DetectCollisions(t *testing.T) {
	e := createTestEngine()

	tests := []struct {
		name         string
		ballPos      Vec2
		p1Pos        Vec2
		p2Pos        Vec2
		expectedColl engine.Collision
	}{
		{
			name:         "no collision - ball in center",
			ballPos:      Vec2{X: 400, Y: 200},
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollNone,
		},
		{
			name:         "collision with P1",
			ballPos:      Vec2{X: 55, Y: 160}, // Close to P1
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollP1Top, // The ball position triggers top collision, not center
		},
		{
			name:         "collision with P2",
			ballPos:      Vec2{X: 735, Y: 160}, // Close to P2
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollP2Top, // The ball position triggers top collision, not center
		},
		{
			name:         "collision with top wall",
			ballPos:      Vec2{X: 400, Y: -5},
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollTop,
		},
		{
			name:         "collision with bottom wall",
			ballPos:      Vec2{X: 400, Y: 395},
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollBottom,
		},
		{
			name:         "collision with left wall (P2 wins)",
			ballPos:      Vec2{X: -10, Y: 200},
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollLeft,
		},
		{
			name:         "collision with right wall (P1 wins)",
			ballPos:      Vec2{X: 810, Y: 200},
			p1Pos:        Vec2{X: 50, Y: 160},
			p2Pos:        Vec2{X: 740, Y: 160},
			expectedColl: engine.CollRight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e.BallPos = tt.ballPos
			e.P1Pos = tt.p1Pos
			e.P2Pos = tt.p2Pos

			collision := e.detectColl()
			assert.Equal(t, tt.expectedColl, collision)
		})
	}
}

func TestCanvasEngine_PaddleMovement(t *testing.T) {
	e := createTestEngine()
	initialP1Pos := e.P1Pos
	initialP2Pos := e.P2Pos

	// Test P1 movement
	e.p1Up()
	assert.True(t, e.P1Vel.Y < 0, "P1 should move up (negative Y velocity)")

	e.p1Down()
	assert.True(t, e.P1Vel.Y > 0, "P1 should move down (positive Y velocity)")

	// Test P2 movement
	e.p2Up()
	assert.True(t, e.P2Vel.Y < 0, "P2 should move up (negative Y velocity)")

	e.p2Down()
	assert.True(t, e.P2Vel.Y > 0, "P2 should move down (positive Y velocity)")

	// Verify positions haven't changed yet (movement happens in tick)
	assert.Equal(t, initialP1Pos, e.P1Pos)
	assert.Equal(t, initialP2Pos, e.P2Pos)
}

func TestCanvasEngine_BallVelocityInversion(t *testing.T) {
	e := createTestEngine()

	// Set initial ball velocity
	e.BallVel = Vec2{X: 2.0, Y: 1.5}
	initialVelX := e.BallVel.X
	initialVelY := e.BallVel.Y

	// Test X velocity inversion
	e.inverseBallXVelocity()
	assert.Equal(t, -initialVelX, e.BallVel.X)
	assert.Equal(t, initialVelY, e.BallVel.Y) // Y should remain unchanged

	// Test Y velocity inversion
	e.inverseBallYVelocity()
	assert.Equal(t, -initialVelX, e.BallVel.X) // X should remain unchanged
	assert.Equal(t, -initialVelY, e.BallVel.Y)
}

func TestCanvasEngine_BallAdvancement(t *testing.T) {
	e := createTestEngine()

	// Set initial position and velocity
	initialPos := Vec2{X: 100, Y: 100}
	velocity := Vec2{X: 2.0, Y: 1.0}
	e.BallPos = initialPos
	e.BallVel = velocity

	// Store initial velocity multiplier and FPS for calculation
	initialVelMultiplier := e.VelocityMultiplier

	// Advance ball
	e.advanceBall()

	// Ball movement calculation:
	// velocityMultiplier increases by DEFAULT_VEL_INCR (0.0005)
	// dt = 1.0 / e.FPS
	// velocityThisTick = velocity * (initialVelMultiplier + DEFAULT_VEL_INCR) * dt
	// newPos = initialPos + velocityThisTick

	expectedVelMultiplier := initialVelMultiplier + DEFAULT_VEL_INCR
	dt := 1.0 / e.FPS
	velocityThisTick := velocity.Scale(expectedVelMultiplier * dt)
	expectedPos := initialPos.Add(velocityThisTick)

	assert.InDelta(t, expectedPos.X, e.BallPos.X, 0.001)
	assert.InDelta(t, expectedPos.Y, e.BallPos.Y, 0.001)
}

func TestCanvasEngine_Reset(t *testing.T) {
	e := createTestEngine()

	// Change some values
	e.P1Score = 5
	e.P2Score = 3
	e.BallPos = Vec2{X: 100, Y: 100}
	e.BallVel = Vec2{X: 5, Y: 5}
	e.P1Pos = Vec2{X: 100, Y: 100}
	e.P2Pos = Vec2{X: 100, Y: 100}

	// Reset
	e.reset()

	// Scores should remain (reset doesn't clear scores)
	assert.Equal(t, 5, e.P1Score)
	assert.Equal(t, 3, e.P2Score)

	// Ball and paddle positions should be reset
	// We can't test exact values without knowing the reset logic,
	// but we can verify they changed from the test values
	assert.NotEqual(t, Vec2{X: 100, Y: 100}, e.BallPos)
}

func TestCanvasEngine_ResetBall(t *testing.T) {
	e := createTestEngine()

	// Change ball state
	e.BallPos = Vec2{X: 100, Y: 100}
	e.BallVel = Vec2{X: 5, Y: 5}

	// Reset ball
	e.resetBall()

	// Ball should be repositioned and velocity should be reset
	// Exact values depend on reset logic, but position should change
	assert.NotEqual(t, Vec2{X: 100, Y: 100}, e.BallPos)
	assert.NotEqual(t, Vec2{X: 5, Y: 5}, e.BallVel)
}

func TestCanvasEngine_OutOfBounds(t *testing.T) {
	e := createTestEngine()

	// Test paddle out of bounds correction
	e.P1Pos = Vec2{X: 0, Y: -100}  // Above screen
	e.P2Pos = Vec2{X: 790, Y: 500} // Below screen

	e.deOutOfBoundsPlayers()

	// Paddles should be constrained within screen bounds
	assert.True(t, e.P1Pos.Y >= 0, "P1 should not be above screen")
	assert.True(t, e.P2Pos.Y <= e.Game.Height-e.Game.P2.Height, "P2 should not be below screen")

	// Test ball out of bounds correction
	e.BallPos = Vec2{X: -100, Y: -100}
	e.deOutOfBoundsBall()

	// Ball position should be corrected (exact behavior depends on implementation)
	// We just verify the function doesn't panic and potentially corrects position
}

func TestVec2_Operations(t *testing.T) {
	v1 := Vec2{X: 3, Y: 4}
	v2 := Vec2{X: 1, Y: 2}

	// Test Add
	result := v1.Add(v2)
	assert.Equal(t, Vec2{X: 4, Y: 6}, result)

	// Test Scale
	scaled := v1.Scale(2.0)
	assert.Equal(t, Vec2{X: 6, Y: 8}, scaled)

	// Test Scale with negative
	scaledNeg := v1.Scale(-0.5)
	assert.Equal(t, Vec2{X: -1.5, Y: -2}, scaledNeg)
}
