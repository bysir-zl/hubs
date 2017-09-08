package tests

import (
	"testing"
	"bytes"
	"time"
	"log"
	"io"
	"encoding/binary"
)

func TestBinary(t *testing.T) {
	bs:=[]byte{}
	bf:=bytes.NewBuffer(bs)
	
	go func() {
		var l uint8 = 4
		binary.Write(bf,binary.LittleEndian,&l)
		
		bf.Write([]byte("hh"))
	}()
	
	go func() {
		var ll uint8
		binary.Read(bf,binary.LittleEndian,&ll)

		log.Print(ll,bf.Bytes())

		x:=make([]byte,2)
		i,err:=io.ReadFull(bf,x)
		log.Print(i,err,string(x))


	}()
		
	time.Sleep(1*time.Hour)
	
}

func TestMove(t *testing.T) {
	t.Log(1<<(2*4))
}
