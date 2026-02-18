package domain

import (
	"time"

	spotpb "github.com/chilly266futon/exchange-service-contracts/gen/pb/spot"
)

type Market struct {
	ID           string
	Name         string
	Description  string
	Enabled      bool
	DeletedAt    *time.Time
	AllowedRoles []spotpb.UserRole
}

func (m *Market) IsAvailable() bool {
	return m.Enabled && m.DeletedAt == nil
}

func (m *Market) IsAccessibleForRoles(userRoles []spotpb.UserRole) bool {
	if len(m.AllowedRoles) == 0 {
		return true
	}

	allowedMap := make(map[spotpb.UserRole]bool, len(m.AllowedRoles))
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
