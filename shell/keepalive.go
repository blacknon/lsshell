// Copyright (c) 2024 Blacknon. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package shell

import (
	"fmt"
	"os"
	"sync"
)

func (s *shell) checkKeepalive() {
	result := []*sConnect{}
	ch := make(chan bool)
	m := new(sync.Mutex)
	clients := s.Connects

	for _, client := range clients {
		go func(client *sConnect) {
			// keepalive
			err := client.Connect.CheckClientAlive()

			if err != nil {
				// error
				fmt.Fprintf(os.Stderr, "Exit Connect %s, Error: %s\n", client.Name, err)

				// close sftp client
				client.Client.Close()
			} else {
				// delete client from map
				m.Lock()
				result = append(result, client)
				m.Unlock()
			}

			ch <- true
		}(client)
	}

	// wait
	for i := 0; i < len(clients); i++ {
		<-ch
	}

	s.Connects = result

	if len(clients) == 0 {
		s.exit(1, "Error: No valid connections\n")
	}

	return
}
