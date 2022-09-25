package abi

import (
	"log"
	"testing"
)

func TestEncodeSimple(t *testing.T) {
	b := NewBuilder(nil)
	b.WriteInt(1001)
	b.WriteInt(1234)
	b.WriteInt(1555)
	ans := b.Finish()
	log.Println(PrettyHex(ans))
}

func TestEncodeDynamic(t *testing.T) {
	b := NewBuilder(nil)
	ans := b.
		EnterDynamic(2).
		WriteInt(123).WriteInt(123).
		ExitDynamic().
		WriteInt(4414).
		Finish()
	log.Println(PrettyHex(ans))
}
func TestEncodeDynamicComplex(t *testing.T) {
	b := NewBuilder(nil)
	ans := b.
		EnterDynamic(2).
		EnterDynamic(2).WriteInt(1).WriteInt(2).ExitDynamic().
		EnterDynamic(1).WriteInt(3).ExitDynamic().
		ExitDynamic().
		EnterDynamic(3).
		WriteString("one").WriteString("two").WriteString("three").
		ExitDynamic().
		Finish()
	log.Println(PrettyHex(ans))
}

func TestEncodeString(t *testing.T) {
	b := NewBuilder(nil)
	ans := b.
		WriteString("hello!").
		WriteInt(4414).
		Finish()
	log.Println(PrettyHex(ans))
}

func TestEncodeLongString(t *testing.T) {
	b := NewBuilder(nil)
	ans := b.
		WriteString("hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello! hello!  ").
		WriteInt(4414).
		Finish()
	log.Println(PrettyHex(ans))
}

func TestEncodeNestedDynamic(t *testing.T) {
	b := NewBuilder(nil)
	ans := b.
		EnterDynamic(4).
		WriteString("hello!").
		WriteString("hello?").
		WriteString("hello.").
		WriteString("hello,").
		ExitDynamic().
		WriteInt(4414).
		Finish()
	log.Println(PrettyHex(ans))
}
