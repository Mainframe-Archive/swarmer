package util

import "net"

type ILookup interface {
	GetIP(domain string) (string, error)
}

type Lookup struct {
	ILookup
}

func GetLookup() *Lookup {
	var l = Lookup{}

	return &l
}

func (l *Lookup) GetIP(domain string) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}

	ip := ips[0].String()

	return ip, nil
}
