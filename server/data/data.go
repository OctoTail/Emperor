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
var Timers map[uint32]*Timer
var timersI uint32
var fovUnits map[uint32]map[uint32]bool
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
	//MvTimer aUint32
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
	/*unit.MvTimer.class = 0
	unit.MvTimer.id = id
	unit.MvTimer.field = 6*/

	return &unit
}

type Struct struct {
	sync.Mutex
}

type Timer struct {
	inform map[uint16]uint
	id uint32
	Start float64
	Delta float64
	fn func(args ...interface{})
	args []interface{}
	Dead bool
}

func (self *Timer) Go() {
	self.Start = time.Since(RefT).Seconds()
	for k, v := range (self.inform) {
		if (v == 0) {
			continue
		}
		msg := make([]byte, 21)
		msg[0] = 1 //Class
		binary.BigEndian.PutUint32(msg[1:5], self.id)
		binary.BigEndian.PutUint64(msg[5:13], math.Float64bits(self.Start))
		binary.BigEndian.PutUint64(msg[13:21], math.Float64bits(self.Delta))
		res := encryption.Encrypt(msg, Players[k].Key)
		conn.WriteTo(res, Players[k].Addr)
	}
	go func(){
		time.Sleep(time.Duration(float64(time.Second)*self.Delta)) //Damn casting
		if self.Dead {
			return
		} else {
			self.fn(self.args...)
		}
	} ()
}

func (self *Timer) Die() {
	self.Dead = true
	for k, v := range (self.inform) {
		if (v == 0) {
			continue
		}
		msg := make([]byte, 21)
		msg[0] = 1 //Class
		binary.BigEndian.PutUint32(msg[1:5], self.id)
		res := encryption.Encrypt(msg, Players[k].Key)
		conn.WriteTo(res, Players[k].Addr)
	}
}

type a struct {
	inform map[uint16]uint
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
	for k, v := range self.inform {
		if (v == 0) {
			continue
		}
        self.send_inform(k)
	}
}

func (self *aUint32) SetInform(pId uint16) {
    self.inform[pId]++
    if self.inform[pId] == 1 {
        self.send_inform(pId)
    }
}

func (self *aUint32) send_inform(pId uint16) {
    msg := make([]byte, 10)
    msg[0] = self.class
    binary.BigEndian.PutUint32(msg[1:5], self.id)
    msg[5]=self.field
    binary.BigEndian.PutUint32(msg[6:10], self.A)
    res := encryption.Encrypt(msg, Players[pId].Key)
    conn.WriteTo(res, Players[pId].Addr)
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
	for k, v := range self.inform {
		if (v == 0) {
			continue
		}
        self.send_inform(k)
	}
}

func (self *aFloat64) SetInform(pId uint16) {
    self.inform[pId]++
    if self.inform[pId] == 1 {
        self.send_inform(pId)
    }
}

func (self *aFloat64) send_inform(pId uint16) {
    msg := make([]byte, 14)
    msg[0] = self.class
    binary.BigEndian.PutUint32(msg[1:5], self.id)
    msg[5]=self.field
    binary.BigEndian.PutUint64(msg[6:14], math.Float64bits(self.A))
    res := encryption.Encrypt(msg, Players[pId].Key)
    conn.WriteTo(res, Players[pId].Addr)
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
	Timers = make(map[uint32]*Timer)
	fovUnits = make(map[uint32]map[uint32]bool)
	timersI = 0
	Players[0] = &Player{Name:"root",
		Key:[]byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16}}

	Units[0] = NewUnit(0)
	Units[0].Owner.A = 0
	Units[0].MaxSpeed.A = 10
	Units[0].Vx.inform = map[uint16]uint{0:1}
	Units[0].Vy.inform = map[uint16]uint{0:1}
	Units[0].X.inform = map[uint16]uint{0:1}
	Units[0].Y.inform = map[uint16]uint{0:1}
	//Units[0].MvTimer.inform = map[uint16]uint{0:1}

	Units[1] = NewUnit(1)
	Units[1].Owner.A = 1
	Units[1].MaxSpeed.A = 10
	Units[1].Vx.inform = map[uint16]uint{1:1}
	Units[1].Vy.inform = map[uint16]uint{1:1}
	Units[1].X.inform = map[uint16]uint{1:1}
	Units[1].Y.inform = map[uint16]uint{1:1}
	//Units[1].MvTimer.inform = map[uint16]uint{1:1}

    for i := range Units {
        fovUnits[i] = make(map[uint32]bool)
    }

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
			x := math.Float64frombits(binary.BigEndian.Uint64(msg[6:14]))
			y := math.Float64frombits(binary.BigEndian.Uint64(msg[14:22]))
			unit.X.Update(x)
			unit.Y.Update(y)


		case 1: //Vx,Vy
			vx := math.Float64frombits(binary.BigEndian.Uint64(msg[6:14]))
			vy := math.Float64frombits(binary.BigEndian.Uint64(msg[14:22]))
			if math.Sqrt(vx*vx + vy*vy) > unit.MaxSpeed.A {
				return errors.New("Desired speed > Unit.MaxSpeed")
			}
			deltafov(unitId, 5, vx, vy)
			unit.Vx.Update(vx)
			unit.Vy.Update(vy)
			updateFn := func (id ...interface{}) {
				unit := Units[id[0].(uint32)]
				unit.Lock()
				defer unit.Unlock()
				//TODO: Check for confilcts
				unit.X.Update(unit.X.A+unit.Vx.A)
				unit.Y.Update(unit.Y.A+unit.Vy.A)}
			Timers[timersI] = &Timer{fn:updateFn,
				args:[]interface{}{unitId},
				Delta:1.0,
				id:timersI,
				inform:unit.Vx.inform}
			Timers[timersI].Go()
			//unit.MvTimer.Update(timersI)
			timersI++

		}
	}
	return nil
}

func deltafov(uId uint32, r, dx, dy float64) { //TODO
	unit := Units[uId]
	for i, u := range Units {
		if fovUnits[uId][i] {
			continue
		}
		if math.Pow(u.X.A-unit.X.A-dx, 2) + math.Pow(u.Y.A-unit.Y.A-dy, 2) <= r {
			fovUnits[uId][i] = true
            u.X.SetInform(unit.Owner.A)
            u.Y.SetInform(unit.Owner.A)
			//I see him now
		}
	}
	for k, v := range fovUnits[uId] {
		if !v {
			continue
		}
		if math.Pow(Units[k].X.A-unit.X.A-dx, 2) + math.Pow(Units[k].Y.A-unit.Y.A-dy, 2) > r {
			fovUnits[uId][k] = false
            Units[k].X.inform[unit.Owner.A]--
            if Units[k].X.inform[unit.Owner.A] <= 0 {
                delete(Units[k].X.inform, unit.Owner.A)
            }
            Units[k].Y.inform[unit.Owner.A]--
            if Units[k].Y.inform[unit.Owner.A] <= 0 {
                delete(Units[k].Y.inform, unit.Owner.A)
            }
			//I don't see him anymore
	    }
    }
}
