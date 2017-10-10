package main

import (
	"github.com/coreos/go-systemd/dbus"
        "fmt"
)

func main() {
  conn, err := dbus.New()
  if err != nil {
  }
  defer conn.Close()

  unitStatuses, err := conn.ListUnits()
  if err != nil {
  }

  for _, unitStatus := range unitStatuses {
    fmt.Printf("%s\n", unitStatus.Name)

    properties, err := conn.GetUnitProperties(unitStatus.Name)

    if err != nil {
    }

    for name, value := range properties {
      value, ok := value.(uint64)

      if ok {
        fmt.Printf("%+v=%+v\n", name, value)
      }
    }


  }


}
