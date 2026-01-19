package domain

import (
	spotv1 "spotService/gen/pb"
	"time"
)

type Market struct {
	ID           string
	Name         string
	Description  string
	Enabled      bool
	DeletedAt    *time.Time
	AllowedRoles []spotv1.UserRole
}

func (m *Market) IsAvailable() bool {
	return m.Enabled && m.DeletedAt == nil
}

func (m *Market) IsAccessibleForRoles(userRoles []spotv1.UserRole) bool {
	if len(m.AllowedRoles) == 0 {
		return true
	}

	allowedMap := make(map[spotv1.UserRole]bool, len(m.AllowedRoles))
	for _, role := range m.AllowedRoles {
		allowedMap[role] = true
	}

	for _, userRole := range userRoles {
		if allowedMap[userRole] {
			return true
		}
	}

	return false
}
