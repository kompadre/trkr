package main

import (
	"fmt"
	"os"
)

func main() {
	syx, err := os.ReadFile("./assets/syx/Dexed_01.syx")
	if err != nil {
		panic(err)
	}
	start := 0x7c
	interval := 0xf0 - 0x70
	length := 0x85 - start
	for i := start; i < len(syx); i += interval {
		fmt.Printf("%x:%s\n", i, string(syx[i:i+length]))
	}

}
