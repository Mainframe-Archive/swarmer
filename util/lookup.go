package util

import "net"

// ILookup is the interface for interacting with IPLookup services.
type ILookup interface {
	GetIP(domain string) (string, error)
}

// Lookup is the struct for this implementation of ILookup.
type Lookup struct {
	ILookup
}

// GetLookup returns a pointer to an instance of this implementation of ILookup.
func GetLookup() *Lookup {
	var l = Lookup{}

	return &l
}

// GetIP takes a domain name and returns the DNS results for the IP of the given domain.
func (l *Lookup) GetIP(domain string) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}

	ip := ips[0].String()

	return ip, nil
}
