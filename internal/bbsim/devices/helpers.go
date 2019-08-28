package devices

import "github.com/looplab/fsm"

var newFSM = fsm.NewFSM

func getOperStateFSM(cb fsm.Callback) *fsm.FSM {
	return newFSM(
		"down",
		fsm.Events{
			{Name: "enable", Src: []string{"down"}, Dst: "up"},
			{Name: "disable", Src: []string{"up"}, Dst: "down"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				cb(e)
			},
		},
	)
}
