package jiragoli

import (
	"fmt"
	"strings"
)

type JIRAUser struct {
	Name    string `json:"name"`
	Email   string `json:"emailAddress"`
	UserURL string `json:"self"`
}

func (ju JIRAUser) String() string {
	return strings.Trim(fmt.Sprintf("%s (%s)", strings.Title(ju.Name), ju.Email), " ")
}
