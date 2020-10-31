package main

import (
	"github.com/vishvananda/netlink"
	"log"
)

func main() {

	ll, err := netlink.LinkList()
	if err != nil {
		log.Panic(err)
	}
	for _, l := range ll {
		rs, _ := netlink.RouteList(l, netlink.FAMILY_ALL)
		log.Print(rs)
		log.Print(l.Type(), l.Attrs())
	}
	rs, _ := netlink.RouteList(nil, netlink.FAMILY_ALL)
	log.Print(rs)

}
