package xep0045

import (
	"testing"
	"fmt"
)

func TestXEP0045_RoomConfig(t *testing.T) {
	roomComfig := NewRoomConfig()
	formElem := roomComfig.AsElement()
	fmt.Println(formElem.String())

}
