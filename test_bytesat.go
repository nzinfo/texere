package main

import "fmt"

func main() {
    // Test edge cases
    cases := []struct{
        name string
        size int
        byteIdx int
    }{
        {"first byte", 10, 0},
        {"last byte", 10, 9},
        {"middle byte", 10, 5},
        {"single byte", 1, 0},
    }
    
    for _, c := range cases {
        fmt.Printf("%s: size=%d, byteIdx=%d, byteIdx-1=%d\n", 
            c.name, c.size, c.byteIdx, c.byteIdx-1)
    }
}
