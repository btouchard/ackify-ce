// SPDX-License-Identifier: AGPL-3.0-or-later
package models

import "strings"

type User struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (u *User) IsValid() bool {
	return strings.TrimSpace(u.Sub) != "" && strings.TrimSpace(u.Email) != ""
}

func (u *User) NormalizedEmail() string {
	return strings.ToLower(u.Email)
}
