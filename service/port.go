package service

import "net"

// Returns the next free TCP port. Otherwise,
// return an error.
//
// This function tries to create a connection on localhost:0.
// If it can, that means the port is free. So return the stored
// port number back to the user.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return -1, err
	}

	l, err := net.ListenTCP("tcp", addr)
	defer l.Close()
	if err != nil {
		return -1, err
	}

	return l.Addr().(*net.TCPAddr).Port, nil
}
