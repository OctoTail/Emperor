package data

import (
	"sync"
	"time"
	"errors"
	"encoding/binary"
	"math"
	"net"
	"../encryption"
)

var Players map[uint16]*Player
var Units map[uint32]*Unit
var Structs map[uint32]*Struct
var RefT time.Time
var conn *net.UDPConn

type Player struct {
	Mac [6]byte
	Name string
	Email string
	Key []byte
	Addr net.Addr
}

type City struct {
	sync.Mutex
	Name aString
	Pop aUint32
	Structs aStructs
	Units aUnits
}

type Unit struct {
	sync.Mutex
	Owner aUint16
	X aFloat64
	Y aFloat64
	Vx aFloat64
	Vy aFloat64
	MaxSpeed aFloat64
}

func NewUnit(id uint32) *Unit {
	unit := Unit{}
	unit.Owner.class = 0
	unit.Owner.id = id
	unit.Owner.field = 0
	unit.X.class = 0
	unit.X.id = id
	unit.X.field = 1
	unit.Y.class = 0
	unit.Y.id = id
	unit.Y.field = 2
	unit.Vx.class = 0
	unit.Vx.id = id
	unit.Vx.field = 3
	unit.Vy.class = 0
	unit.Vy.id = id
	unit.Vy.field = 4
	unit.MaxSpeed.class = 0
	unit.MaxSpeed.id = id
	unit.MaxSpeed.field = 5

	return &unit
}

type Struct struct {
	sync.Mutex
}

type Timer struct {
	Start float64
	Delta float64
	fn func(args ...interface{})
	args []interface{}
	Dead bool
}

func (self Timer) Go() {
	go func(){
		self.Start = time.Since(RefT).Seconds()
		time.Sleep(time.Duration(float64(time.Second)*self.Delta)) //Damn casting
		if self.Dead {
			return
		} else {
			self.fn(self.args...)
		}
	} ()
}

type a struct {
	inform []uint16
	class byte
	id uint32
	field byte
}

type aString struct {
	a
	A string
}
func (self *aString) Update(value string) {
	self.A = value
/*	for i := range inform {

	}*/
}

type aUint32 struct {
	a
	A uint32
}
func (self *aUint32) Update(value uint32) {
	self.A = value
/*	for i := range inform {

	}*/
}

type aUint16 struct {
	a
	A uint16
}
func (self *aUint16) Update(value uint16) {
	self.A = value
/*	for i := range inform {

	}*/
}

type aFloat64 struct {
	a
	A float64
}

func (self *aFloat64) Update(value float64) {
	self.A = value
	for _, i := range self.inform {
		msg := make([]byte, 14)
		msg[0] = self.class
		binary.BigEndian.PutUint32(msg[1:5], self.id)
		msg[5]=self.field
		binary.BigEndian.PutUint64(msg[6:14], math.Float64bits(value))
		res := encryption.Encrypt(msg, Players[i].Key)
		conn.WriteTo(res, Players[i].Addr)
	}
}

type aPtrs struct {
	A []uint32
	inform []uint16
}
func (self *aPtrs) Delete(index int) {
	self.A[index] = self.A[len(self.A)-1] 
	self.A = self.A[:len(self.A)-1]
	self.update()
}
func (self *aPtrs) Add(item uint32) {
	self.A = append(self.A, item)
	self.update() 
}
func (self *aPtrs) update() {
	/*	for i := range inform {

	}*/
}

type aStructs aPtrs
type aUnits aPtrs

func Init(sCon *net.UDPConn) {
	conn = sCon
	Players = make(map[uint16]*Player)
	Units = make(map[uint32]*Unit)
	Structs = make(map[uint32]*Struct)
	Players[0] = &Player{Name:"root",
		Key:[]byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16}}

	Units[0] = NewUnit(0)
	Units[0].Owner.A = 0
	Units[0].MaxSpeed.A = 10
	Units[0].Vx.inform = []uint16{0}
	Units[0].Vy.inform = []uint16{0}

	Structs[0] = &Struct{}

	RefT = time.Now()
}

func ReqSET2(msg []byte, pId uint16) error {
	// TODO: check len(msg)
	switch msg[0] {
	case 0: //Unit
		unitId := binary.BigEndian.Uint32(msg[1:5])
		unit, ok := Units[unitId]
		if !ok {
			return errors.New("No such Unit")
		}
		unit.Lock()
		defer unit.Unlock()
		if unit.Owner.A != pId {
			return errors.New("Do not own Unit")
		}
		switch msg[5] { //Switch attributes
		case 0: //X,Y
			if pId != 0 {
				return errors.New("Can't change Unit.X/Y")
			}

		case 1: //Vx,Vy
			vx := math.Float64frombits(binary.BigEndian.Uint64(msg[6:14]))
			vy := math.Float64frombits(binary.BigEndian.Uint64(msg[14:22]))
			if math.Sqrt(vx*vx + vy*vy) > unit.MaxSpeed.A {
				return errors.New("Desired speed > Unit.MaxSpeed")
			}
			unit.Vx.Update(vx)
			unit.Vy.Update(vy)
		}
	}
	return nil
}