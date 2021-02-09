/*
 * Copyright (c) The One True Way 2021. Apache License 2.0. The authors accept no liability, 0 nada for the use of this software.  It is offered "As IS"  Have fun with it!!
 */

package main

import (
	"fmt"
	"time"
)

func main(){
	d:=0
	for true{
		fmt.Printf("Hello %d \n",d)
		time.Sleep(2*time.Second)
		d=d+1
	}
}
