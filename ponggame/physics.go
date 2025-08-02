//go:build !prototype
// +build !prototype

package ponggame

import (
	"math"

	"github.com/ndabAP/ping-pong/engine"
)

// -----------------------------------------------------------------------------
// Paper‑driven AVBD parameters (SIGGRAPH 25, Table 2)
// -----------------------------------------------------------------------------
const (
	DEFAULT_FPS      = 60
	DEFAULT_VEL_INCR = 0.05

	kStart       = 1.0e-3      // initial stiffness for all hard constraints
	beta         = 10.0        // stiffness growth factor
	alpha        = 0.95        // dual‑variable damping
	gamma        = 0.99        // stiffness damping
	cor          = 0.9         // 0=inélastico, 1=elástico perfeito
	maxAngle     = math.Pi / 3 // 60°
	minSpeed     = 0.08        // fração da largura/seg   (saque)
	maxSpeed     = 1.30        // fração da largura/seg   (teto absoluto)
	minWallAngle = 0.15

	canvas_border_correction = 1

	initial_ball_x_vel = 0.1
	initial_ball_y_vel = 0.1

	y_vel_ratio = 1

	maxVertices    = 32 // 4 wall verts + 4 ball verts + 2 paddle verts
	maxConstraints = 128
)

// -----------------------------------------------------------------------------
// Construction helpers
// -----------------------------------------------------------------------------

func (p *AVBDPhysics) addVertex(pos Vec2, invMass, radius float64) int {
	v := Vertex{Pos: pos, Prev: pos, InvMass: invMass, Radius: radius}
	p.verts = append(p.verts, v)
	return len(p.verts) - 1
}

// soft XPBD‑style distance spring
func (p *AVBDPhysics) addDistConstraint(i, j int, rest float64) {
	p.cons = append(p.cons, Constraint{I: i, J: j, Rest: rest, Hard: false, K: 1 / rest})
	p.staticCount++
}

// one‑sided plane contact (half‑space)
func (p *AVBDPhysics) addPlaneContact(v int, n Vec2, d float64) {
	p.cons = append(p.cons, Constraint{
		I: v, J: -1,
		N: n, Radius: d,
		Hard: true, K: kStart,
		MinLam: 0, MaxLam: math.Inf(1),
	})
}

// -----------------------------------------------------------------------------
// Public constructor – builds permanent geometry
// -----------------------------------------------------------------------------

func NewAVBDPhysics(e *CanvasEngine) *AVBDPhysics {
	p := &AVBDPhysics{gravity: Vec2{0, 0}, damping: 0.999}
	p.verts = make([]Vertex, 0, maxVertices)
	p.cons = make([]Constraint, 0, maxConstraints)

	// ----- static wall marker vertices (pinned) -----------------------------
	lw := p.addVertex(Vec2{0, e.Game.Height * 0.5}, 0, 0)
	rw := p.addVertex(Vec2{e.Game.Width, e.Game.Height * 0.5}, 0, 0)
	tw := p.addVertex(Vec2{e.Game.Width * 0.5, 0}, 0, 0)
	bw := p.addVertex(Vec2{e.Game.Width * 0.5, e.Game.Height}, 0, 0)
	_ = []int{lw, rw, tw, bw} // silence lints – positions are read later

	// ----- ball as a 4‑corner square ---------------------------------------
	ballR := e.Game.Ball.Height * 0.5
	cx := e.BallPos.X + ballR
	cy := e.BallPos.Y + ballR

	b0 := p.addVertex(Vec2{cx - ballR, cy - ballR}, 1, ballR)
	b1 := p.addVertex(Vec2{cx + ballR, cy - ballR}, 1, ballR)
	b2 := p.addVertex(Vec2{cx + ballR, cy + ballR}, 1, ballR)
	b3 := p.addVertex(Vec2{cx - ballR, cy + ballR}, 1, ballR)

	stiff := 0.0001
	p.addDistConstraint(b0, b1, 2*ballR)
	p.addDistConstraint(b1, b2, 2*ballR)
	p.addDistConstraint(b2, b3, 2*ballR)
	p.addDistConstraint(b3, b0, 2*ballR)
	diag := 2 * ballR * math.Sqrt2
	p.addDistConstraint(b0, b2, diag*stiff)
	p.addDistConstraint(b1, b3, diag*stiff)

	// ----- paddle endpoints (pinned, kinematic) ----------------------------
	p1h := e.Game.P1.Height * 0.5
	p1cx := e.P1Pos.X + e.Game.P1.Width*0.5
	p1cy := e.P1Pos.Y + p1h
	_ = p.addVertex(Vec2{p1cx, p1cy - p1h}, 0, 0)
	_ = p.addVertex(Vec2{p1cx, p1cy + p1h}, 0, 0)

	p2cx := e.P2Pos.X + e.Game.P2.Width*0.5
	p2cy := e.P2Pos.Y + e.Game.P2.Height*0.5
	_ = p.addVertex(Vec2{p2cx, p2cy - p1h}, 0, 0)
	_ = p.addVertex(Vec2{p2cx, p2cy + p1h}, 0, 0)

	return p
}

// -----------------------------------------------------------------------------
// Per‑frame contact generation
// -----------------------------------------------------------------------------

func (p *AVBDPhysics) makeContacts(e *CanvasEngine) {
	// trim contacts back to the static set
	p.cons = p.cons[:p.staticCount]

	w := e.Game.Width
	h := e.Game.Height

	for _, bi := range []int{4, 5, 6, 7} { // ball vertices
		v := &p.verts[bi].Pos

		// walls
		if v.X < 0 {
			p.addPlaneContact(bi, Vec2{1, 0}, 0)
		}
		if v.X > w {
			p.addPlaneContact(bi, Vec2{-1, 0}, -w)
		}
		if v.Y < 0 {
			p.addPlaneContact(bi, Vec2{0, 1}, 0)
		}
		if v.Y > h {
			p.addPlaneContact(bi, Vec2{0, -1}, -h)
		}

		// paddles treated as vertical planes
		// P1
		paddW := e.Game.P1.Width
		if v.X < e.P1Pos.X+paddW && v.Y > e.P1Pos.Y && v.Y < e.P1Pos.Y+e.Game.P1.Height {
			d := e.P1Pos.X + paddW
			p.addPlaneContact(bi, Vec2{1, 0}, d)
		}
		// P2
		if v.X > e.P2Pos.X && v.Y > e.P2Pos.Y && v.Y < e.P2Pos.Y+e.Game.P2.Height {
			d := e.P2Pos.X
			p.addPlaneContact(bi, Vec2{-1, 0}, -d)
		}
	}
}

// -----------------------------------------------------------------------------
// Solver – one AVBD step (Sec. 3.3)
// -----------------------------------------------------------------------------

func (p *AVBDPhysics) Step(dt float64) {
	// 0. warm‑start / stiffness decay (Eq. 19)
	for i := range p.cons {
		c := &p.cons[i]
		if !c.Hard {
			continue
		}
		c.K = math.Max(kStart, gamma*c.K)
		c.Lambda *= alpha * gamma
	}

	// 1. store previous position & integrate forces (standard Verlet)
	for i := range p.verts {
		v := &p.verts[i]
		v.Prev = v.Pos
		if v.InvMass == 0 {
			continue
		}
		v.Vel = v.Vel.Add(p.gravity.Scale(dt))
		v.Pos = v.Pos.Add(v.Vel.Scale(dt))
	}

	// 2. (re)generate contacts for this frame
	p.makeContacts(currentEngine) // currentEngine set by StepPhysics wrapper

	// 3. Gauss‑Seidel primal‑dual iterations
	const iters = 2 // enough with warm‑start
	for n := 0; n < iters; n++ {
		for i := range p.cons {
			c := &p.cons[i]
			vi := &p.verts[c.I]

			// plane vs point
			if c.Hard {
				phi := c.N.Dot(vi.Pos) - c.Radius // C(x)
				grad := c.N
				w := vi.InvMass
				lamCand := c.K*phi + c.Lambda
				lamNew := math.Min(c.MaxLam, math.Max(c.MinLam, lamCand))
				deltaLam := lamNew - c.Lambda
				if deltaLam != 0 && w != 0 {
					corr := deltaLam / w
					vi.Pos = vi.Pos.Sub(grad.Scale(corr * vi.InvMass))
				}
				c.Lambda = lamNew
				if phi < 0 { // penetration ⇒ grow stiffness
					c.K += beta * math.Abs(phi)
				}
				continue
			}

			// soft distance spring (XPBD Eq. 16)
			vj := &p.verts[c.J]
			delta := vj.Pos.Sub(vi.Pos)
			dist := delta.Len()
			if dist == 0 {
				continue
			}
			grad := delta.Scale(1 / dist)
			C := dist - c.Rest
			w := vi.InvMass + vj.InvMass
			corr := -(c.K * C) / (w + c.K)
			vi.Pos = vi.Pos.Add(grad.Scale(corr * vi.InvMass))
			vj.Pos = vj.Pos.Sub(grad.Scale(corr * vj.InvMass))
		}
	}

	// 4. rebuild velocities (xₙ − xₙ₋₁)/dt and damping
	for i := range p.verts {
		v := &p.verts[i]
		if v.InvMass == 0 {
			v.Vel = Vec2{}
			continue
		}
		v.Vel = v.Pos.Sub(v.Prev).Scale(1 / dt).Scale(p.damping)
	}

	for _, c := range p.cons {
		if !c.Hard { // só interessa para contatos unilaterais
			continue
		}
		v := &p.verts[c.I]
		vn := c.N.Dot(v.Vel) // velocidade na direção da normal
		if vn < 0 {          // aproximando-se do plano?
			// v' = v - (1+e) (v·n) n  (reflexão com restituição)
			v.Vel = v.Vel.Sub(c.N.Scale((1 + cor) * vn))
		}
	}
}

// ---------- bridge to CanvasEngine -----------------------------------------

var currentEngine *CanvasEngine // transient pointer used inside makeContacts

func StepPhysics(e *CanvasEngine, dt float64) {
	currentEngine = e // for contact generation

	syncVerticesFromEngine(e)
	e.phy.Step(dt)

	// copy ball centroid back
	var cx, cy float64
	for i := 4; i <= 7; i++ {
		cx += e.phy.verts[i].Pos.X
		cy += e.phy.verts[i].Pos.Y
	}
	cx, cy = cx/4, cy/4
	r := e.Game.Ball.Height * 0.5
	e.BallPos = Vec2{cx - r, cy - r}

	// velocity (average of corners)
	var vx, vy float64
	for i := 4; i <= 7; i++ {
		vx += e.phy.verts[i].Vel.X
		vy += e.phy.verts[i].Vel.Y
	}
	e.BallVel = Vec2{vx / 4, vy / 4}
}

// -----------------------------------------------------------------------------
// Synchronise kinematic bodies (ball & paddles before Step)
// -----------------------------------------------------------------------------

func syncVerticesFromEngine(e *CanvasEngine) {
	// ball
	br := e.Game.Ball.Height * 0.5
	cx := e.BallPos.X + br
	cy := e.BallPos.Y + br
	vx, vy := e.BallVel.X, e.BallVel.Y
	rewind := func(i int, x, y float64) {
		v := &e.phy.verts[i]
		v.Pos = Vec2{x, y}
		v.Prev = v.Pos.Sub(Vec2{vx, vy}.Scale(1 / e.FPS))
		v.Vel = Vec2{vx, vy}
	}
	rewind(4, cx-br, cy-br)
	rewind(5, cx+br, cy-br)
	rewind(6, cx+br, cy+br)
	rewind(7, cx-br, cy+br)

	// paddles (indices 8‑11) – position only, zero mass so velocity zeroed
	p1h := e.Game.P1.Height * 0.5
	p1cx := e.P1Pos.X + e.Game.P1.Width*0.5
	p1cy := e.P1Pos.Y + p1h
	e.phy.verts[8].Pos = Vec2{p1cx, p1cy - p1h}
	e.phy.verts[9].Pos = Vec2{p1cx, p1cy + p1h}

	p2cx := e.P2Pos.X + e.Game.P2.Width*0.5
	p2cy := e.P2Pos.Y + e.Game.P2.Height*0.5
	e.phy.verts[10].Pos = Vec2{p2cx, p2cy - p1h}
	e.phy.verts[11].Pos = Vec2{p2cx, p2cy + p1h}
}

func (e *CanvasEngine) applyPaddleBounce(coll engine.Collision) {
	var paddleY, paddleH float64
	var dir int // +1 se bateu no P1 (bola deve ir p/ direita), -1 p/ P2
	switch coll {
	case engine.CollP1, engine.CollP1Top, engine.CollP1Bottom:
		paddleY = e.P1Pos.Y
		paddleH = e.Game.P1.Height
		dir = +1
	case engine.CollP2, engine.CollP2Top, engine.CollP2Bottom:
		paddleY = e.P2Pos.Y
		paddleH = e.Game.P2.Height
		dir = -1
	default:
		return
	}

	// 1. offset vertical relativo ao centro do paddle
	ballCY := e.BallPos.Y + e.Game.Ball.Height*0.5
	paddleCY := paddleY + paddleH*0.5
	offset := (ballCY - paddleCY) / (paddleH * 0.5) // -1‥+1

	// 2. converte em ângulo
	angle := offset * maxAngle

	// 3. módulo da velocidade (mantém “energia” atual)
	speed := math.Hypot(e.BallVel.X, e.BallVel.Y)
	if speed == 0 {
		speed = e.Game.Width * initial_ball_x_vel // qualquer valor mínimo
	}

	if math.Abs(angle) < minWallAngle {
		// obrigue um leve desvio vertical mesmo em batida central
		angle = math.Copysign(minWallAngle, offset) // usa sinal do offset
	}

	// 4. novo vetor: cos = VX, sen = VY
	e.BallVel.X = float64(dir) * math.Cos(angle) * speed
	e.BallVel.Y = math.Sin(angle) * speed
}

// devolve o vetor já limitado a [min, max]
func clampBallSpeed(v Vec2, w float64) Vec2 {
	s := math.Hypot(v.X, v.Y)
	min := minSpeed * w
	max := maxSpeed * w

	if s < 1e-9 { // bola (quase) parada → força mínimo
		return Vec2{min, min * 0.6} // 0.6: ­inclinação inicial leve
	}
	if s < min {
		k := min / s
		return Vec2{v.X * k, v.Y * k}
	}
	if s > max {
		k := max / s
		return Vec2{v.X * k, v.Y * k}
	}
	return v
}

// func bounceWall(e *CanvasEngine, sign float64) { // sign = +1 teto, -1 piso
// 	speed := math.Hypot(e.BallVel.X, e.BallVel.Y)
// 	if speed < 1e-9 {
// 		speed = minSpeed * e.Game.Width
// 	}

// 	// ângulo atual (após inverter Vy)
// 	ang := math.Atan2(e.BallVel.Y, e.BallVel.X)

// 	// força inclinação mínima
// 	if math.Abs(ang) < minWallAngle {
// 		ang = math.Copysign(minWallAngle, ang)
// 	}

// 	// recria o vetor com o mesmo módulo, mas ângulo ≥ minWallAngle
// 	e.BallVel.X = math.Cos(ang) * speed
// 	e.BallVel.Y = math.Sin(ang) * speed

// 	// clamp final para garantir módulo dentro dos limites
// 	e.BallVel = clampBallSpeed(e.BallVel, e.Game.Width)
// }

func bounceWall(e *CanvasEngine, sign int) { // +1 = teto, -1 = piso
	w := e.Game.Width
	speed := math.Hypot(e.BallVel.X, e.BallVel.Y)
	if speed < 1e-9 {
		speed = minSpeed * w
	}

	// ângulo mínimo acima/abaixo do eixo X
	ang := minWallAngle * float64(sign)
	// conserva direção horizontal (VX)
	if e.BallVel.X < 0 {
		ang = math.Pi - ang
	}

	e.BallVel.X = math.Cos(ang) * speed
	e.BallVel.Y = math.Sin(ang) * speed
	e.BallVel = clampBallSpeed(e.BallVel, w)
}
