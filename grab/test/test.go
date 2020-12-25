package main

import (
	"grab/modules/mongodb"
	"log"
	"net"
)
import "grab"

func main()  {
	module := mongodb.Module{}
 	scanner :=	module.NewScanner()
 	flags := module.NewFlags()
 	scanner.Init(flags.(grab.ScanFlags))
 	target := grab.ScanTarget{
 		IP: net.IP{127,0,0,1},
	}
 	status, res, err:= scanner.Scan(target)
 	if err !=nil{
 		log.Println(status,res)
	}else{
		log.Println(err)
	}
}
