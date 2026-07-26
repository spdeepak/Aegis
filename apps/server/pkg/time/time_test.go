package time

import (
	"fmt"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	fmt.Println(Now())
	fmt.Println(time.Now())
}
