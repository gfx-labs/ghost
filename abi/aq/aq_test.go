package aq_test

import (
	"log"
	"testing"

	"gfx.cafe/open/ghost/abi"
	"gfx.cafe/open/ghost/abi/aq"
)

func TestSearch(t *testing.T) {
	b := abi.NewBuilder(nil)
	ans := b.
		EnterDynamic(2).
		EnterDynamic(2).WriteInt(1).WriteInt(2).ExitDynamic().
		EnterDynamic(1).WriteInt(3).ExitDynamic().
		ExitDynamic().
		EnterDynamic(3).
		WriteString("one").WriteString("two").WriteString("three").
		ExitDynamic().
		Finish()
	vw := aq.NewView(ans)
	log.Println(vw.Search(".{1}<>"))
}
