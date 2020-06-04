package riot

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	assert.NoError(t, nil)
}
