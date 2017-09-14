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
		data:=[]byte("h41562435612h")
		var l uint32 = uint32(len(data))
		binary.Write(bf,binary.LittleEndian,&l)
		
		bf.Write(data)
		
		log.Print("writed: ",bf.Bytes())
	}()
	
	go func() {
		var ll uint32
		binary.Read(bf,binary.LittleEndian,&ll)

		log.Print(ll,bf.Bytes())

		x:=make([]byte,ll)
		i,err:=io.ReadFull(bf,x)
		log.Print(i,err,string(x))

		log.Print("readed: ",bf.Bytes())

	}()
		
	time.Sleep(1*time.Hour)
	
}

func TestMove(t *testing.T) {
	t.Log(1<<(2*4))
}

func TestChanClose(t *testing.T) {
	c:=make(chan int,2)
	go func() {
		c<-1
		c<-1
		c<-1
	}()
	
	time.AfterFunc(1*time.Second, func() {
		close(c)
	})
	
	time.Sleep(1*time.Hour)
}