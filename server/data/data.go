package data

import (
	"sync"
	"time"
)

var Players []Player
var Units []Unit
var Structs []Struct
var RefT time.Time

type Player struct {
	Id uint32
	Mac [6]byte
	Name string
	Email string
	Key []byte
	Ip [4]byte
}

type City struct {
	sync.Mutex
	Id uint32
	Name aString
	Pop aUint32
	Structs aStructs
	Units aUnits
}

type Unit struct {
	sync.Mutex
	Id uint32
}

type Struct struct {
	sync.Mutex
	Id uint32
}

type Timer struct {
	Id uint32
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

type aString struct {
	A string
	inform []uint32
}
func (self *aString) Update(value string) {
	self.A = value
/*	for i := range inform {

	}*/
}

type aUint32 struct {
	A uint32
	inform []uint32
}
func (self *aUint32) Update(value uint32) {
	self.A = value
/*	for i := range inform {

	}*/
}

type aPtrs struct {
	A []uint32
	inform []uint32
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

type aStructs struct {
	aPtrs
}
func (self *aStructs) Iter() <-chan *Struct {
    ch := make(chan *Struct)
    go func () {
        for i := 0; i < len(self.A); i++ {
            ch <- &Structs[self.A[i]]
        }
    } ()
    return ch
}

type aUnits struct {
	aPtrs
}
func (self *aUnits) Iter() <-chan *Unit {
    ch := make(chan *Unit)
    go func () {
        for i := 0; i < len(self.A); i++ {
            ch <- &Units[self.A[i]]
        }
    } ()
    return ch
}

func Init() {
	Players = append(Players, Player{1,
		[6]byte{0,0,0,0,0,0},
		"root",
		"root@mail.com",
		[]byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16},
		[4]byte{0,0,0,0}})
	RefT = time.Now()
}