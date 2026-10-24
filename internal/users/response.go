package users

type UserProfileResponse struct {
	ID          string `json:"ID"`
	UserID      string `json:"UserID"`
	DisplayName string `json:"DisplayName"`
	Biography   string `json:"Biography"`
	UpdatedAt   string `json:"UpdatedAt"`
}

type UserWithProfileResponse struct {
	ID        string              `json:"ID"`
	Username  string              `json:"Username"`
	Email     string              `json:"Email"`
	CreatedAt string              `json:"CreatedAt"`
	Profile   UserProfileResponse `json:"Profile"`
}

// ToUserWithProfileResponse expone únicamente los campos públicos del usuario y su perfil.
func ToUserWithProfileResponse(user UserWithProfile) UserWithProfileResponse {
	return UserWithProfileResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		Profile: UserProfileResponse{
			ID:          user.Profile.ID,
			UserID:      user.Profile.UserID,
			DisplayName: user.Profile.DisplayName,
			Biography:   user.Profile.Biography,
			UpdatedAt:   user.Profile.UpdatedAt,
		},
	}
}
