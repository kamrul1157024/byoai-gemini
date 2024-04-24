package apperror

import "fmt"

func CheckAndPanic(e error) {
    if e != nil {
        panic(e)
    }
}

func CheckAndLog(e error, log *string){
  if e != nil {
    fmt.Println(log, e)
  }
}
