package main

import (
    "semprini/steno/steno/kludge"
    "fmt"
)


func main() {


    foo,err := kludge.DataPath()
    if err != nil {
        fmt.Printf("DataPath() failed: %s\n",err)
    } else {
        fmt.Println(foo)
    }

    foo,err = kludge.PerUserPath()
    if err != nil {
        fmt.Printf("PerUserPath() failed: %s\n",err)
    } else {
        fmt.Println(foo)
    }

}


