package wsclient

import (
	"night-fury/pkgs/log"
	"night-fury/pkgs/utils"
	"night-fury/ws_client/client"
	"sync"
)

func RunClient(addr string, closeChan chan struct{}) error {
	c, err := client.NewClient(addr, closeChan)
	if err != nil {
		log.Errorf(log.TagWSClient, "create new client error %s", err)
		return err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go utils.SafeRun(nil, func() {
		c.ReadMsg()
	})
	go utils.SafeRun(nil, func() {
		defer wg.Done()
		c.WriteMsg()
	})

	wg.Wait()

	return nil
}
