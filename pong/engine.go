package canvas

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"

	"github.com/ndabAP/ping-pong/engine"
)

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
	Debug bool
}

var engineLogger = log.New(os.Stdout, "[ENGINE] ", 0)

// New returns a new Canvas engine for browsers with Canvas support
func New(g engine.Game) *CanvasEngine {
	e := new(CanvasEngine)
	e.Game = g
	e.FPS = DEFAULT_FPS
	e.TPS = 1000.0 / e.FPS

	return e
}

// SetDebug sets the Canvas engines debug state
func (e *CanvasEngine) SetDebug(debug bool) *CanvasEngine {
	engineLogger.Printf("debug %t", debug)
	e.Debug = debug
	return e
}

// SetFPS sets the Canvas engines frames per second
func (e *CanvasEngine) SetFPS(fps uint) *CanvasEngine {
	if fps <= 0 {
		panic("fps must be greater zero")
	}
	engineLogger.Printf("fps %d", fps)
	e.FPS = float64(fps)
	e.TPS = 1000.0 / e.FPS
	return e
}

// Error returns the Canvas engines error
func (e *CanvasEngine) Error() error {
	return e.Err
}

type GameUpdate struct {
	GameWidth     int32   `protobuf:"varint,13,opt,name=gameWidth,proto3" json:"gameWidth,omitempty"`
	GameHeight    int32   `protobuf:"varint,14,opt,name=gameHeight,proto3" json:"gameHeight,omitempty"`
	P1Width       int32   `protobuf:"varint,15,opt,name=p1Width,proto3" json:"p1Width,omitempty"`
	P1Height      int32   `protobuf:"varint,16,opt,name=p1Height,proto3" json:"p1Height,omitempty"`
	P2Width       int32   `protobuf:"varint,17,opt,name=p2Width,proto3" json:"p2Width,omitempty"`
	P2Height      int32   `protobuf:"varint,18,opt,name=p2Height,proto3" json:"p2Height,omitempty"`
	BallWidth     int32   `protobuf:"varint,19,opt,name=ballWidth,proto3" json:"ballWidth,omitempty"`
	BallHeight    int32   `protobuf:"varint,20,opt,name=ballHeight,proto3" json:"ballHeight,omitempty"`
	P1Score       int32   `protobuf:"varint,21,opt,name=p1Score,proto3" json:"p1Score,omitempty"`
	P2Score       int32   `protobuf:"varint,22,opt,name=p2Score,proto3" json:"p2Score,omitempty"`
	BallX         int32   `protobuf:"varint,1,opt,name=ballX,proto3" json:"ballX,omitempty"`
	BallY         int32   `protobuf:"varint,2,opt,name=ballY,proto3" json:"ballY,omitempty"`
	P1X           int32   `protobuf:"varint,3,opt,name=p1X,proto3" json:"p1X,omitempty"`
	P1Y           int32   `protobuf:"varint,4,opt,name=p1Y,proto3" json:"p1Y,omitempty"`
	P2X           int32   `protobuf:"varint,5,opt,name=p2X,proto3" json:"p2X,omitempty"`
	P2Y           int32   `protobuf:"varint,6,opt,name=p2Y,proto3" json:"p2Y,omitempty"`
	P1YVelocity   int32   `protobuf:"varint,7,opt,name=p1YVelocity,proto3" json:"p1YVelocity,omitempty"`
	P2YVelocity   int32   `protobuf:"varint,8,opt,name=p2YVelocity,proto3" json:"p2YVelocity,omitempty"`
	BallXVelocity int32   `protobuf:"varint,9,opt,name=ballXVelocity,proto3" json:"ballXVelocity,omitempty"`
	BallYVelocity int32   `protobuf:"varint,10,opt,name=ballYVelocity,proto3" json:"ballYVelocity,omitempty"`
	Fps           float32 `protobuf:"fixed32,11,opt,name=fps,proto3" json:"fps,omitempty"`
	Tps           float32 `protobuf:"fixed32,12,opt,name=tps,proto3" json:"tps,omitempty"`
	// Optional: if you want to send error messages or debug information
	Error string `protobuf:"bytes,23,opt,name=error,proto3" json:"error,omitempty"`
	Debug bool   `protobuf:"varint,24,opt,name=debug,proto3" json:"debug,omitempty"`
}

// NewRound resets the ball, players and starts a new round. It accepts
// a frames channel to write into and input channel to read from
func (e *CanvasEngine) NewRound(ctx context.Context, framesch chan<- []byte, inputch <-chan []byte, roundResult chan<- int32) {
	engineLogger.Println("new round")

	// time.Sleep(time.Millisecond * 1500) // 1.5 seconds

	e.reset()

	// Calculates and writes frames
	go func() {
		clock := time.NewTicker(time.Duration(e.TPS) * time.Millisecond)
		defer clock.Stop()

		for {
			select {
			case <-ctx.Done():
				engineLogger.Println("exiting")
				return
			case <-clock.C:
				e.tick()

				if errors.Is(e.Err, engine.ErrP1Win) {
					engineLogger.Println("p1 wins")
					e.P1Score += 1

					// Send the winner's ID through the roundResult channel
					select {
					case roundResult <- 1:
					case <-ctx.Done():
						return
					}

					return
				} else if errors.Is(e.Err, engine.ErrP2Win) {
					engineLogger.Println("p2 wins")
					e.P2Score += 1

					// Send the winner's ID through the roundResult channel
					select {
					case roundResult <- 2:
					case <-ctx.Done():
						return
					}

					return
				}

				gameUpdateFrame := GameUpdate{
					GameWidth:     int32(e.Game.Width),
					GameHeight:    int32(e.Game.Height),
					P1Width:       int32(e.Game.P1.Width),
					P1Height:      int32(e.Game.P1.Height),
					P2Width:       int32(e.Game.P2.Width),
					P2Height:      int32(e.Game.P2.Height),
					BallWidth:     int32(e.Game.Ball.Width),
					BallHeight:    int32(e.Game.Ball.Height),
					P1Score:       int32(e.P1Score),
					P2Score:       int32(e.P2Score),
					BallX:         int32(e.BallX),
					BallY:         int32(e.BallY),
					P1X:           int32(e.P1X),
					P1Y:           int32(e.P1Y),
					P2X:           int32(e.P2X),
					P2Y:           int32(e.P2Y),
					P1YVelocity:   int32(e.P1YVelocity),
					P2YVelocity:   int32(e.P2YVelocity),
					BallXVelocity: int32(e.BallXVelocity),
					BallYVelocity: int32(e.BallYVelocity),
					Fps:           float32(e.FPS),
					Tps:           float32(e.TPS),
				}
				jsonTick, err := json.Marshal(gameUpdateFrame)
				if err != nil {
					engineLogger.Printf("Err: %v", err)
				}
				select {
				case framesch <- jsonTick:
					if e.Debug {
						engineLogger.Printf("tick: %s", string(jsonTick))
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Reads user input and moves player one according to it
	go func() {
		for {
			select {
			case key, ok := <-inputch:
				if !ok {
					// Input channel is closed; exit goroutine
					engineLogger.Printf("Input channel closed; exiting input reader goroutine")
					return
				}
				if len(key) == 0 {
					// Empty input received; possibly due to closed channel
					engineLogger.Printf("Received empty input data; exiting")
					return
				}

				in := pong.PlayerInput{}
				err := json.Unmarshal(key, &in)
				if err != nil {
					engineLogger.Printf("Failed to unmarshal input: %v", err)
					// Decide whether to continue or exit; here we'll continue
					continue
				}

				// Process the valid input
				if in.PlayerNumber == int32(1) {
					switch k := in.Input; k {
					case "ArrowUp":
						engineLogger.Printf("key %s", k)
						e.p1Down() // The Canvas origin is top left
					case "ArrowDown":
						engineLogger.Printf("key %s", k)
						e.p1Up()
					}
				} else {
					switch k := in.Input; k {
					case "ArrowUp":
						engineLogger.Printf("key %s", k)
						e.p2Down() // The Canvas origin is top left
					case "ArrowDown":
						engineLogger.Printf("key %s", k)
						e.p2Up()
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
