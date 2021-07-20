package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/dataproviders"
)

func ProcessINodes(channel chan []*dataproviders.INode, store dataproviders.Store, beat Heartbeat) error {
	beat.Start()
	defer beat.Done()
	for {
		select {
		case content := <-channel:
			i := len(content)
			if i == 0 {
				return nil
			}
			store.SaveMany(content...)
			beat.Beat(i)
		default:
		}
	}

}
