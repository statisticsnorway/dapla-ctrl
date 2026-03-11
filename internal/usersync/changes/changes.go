package changes

// UserSyncUserChangeUnit represents a single field change with old and new values.
type UserSyncUserChangeUnit struct {
	Old *string `json:"old"`
	New *string `json:"new"`
}

// UserSyncUserChanges represents a collection of changes made to a user.
type UserSyncUserChanges struct {
	Name        *UserSyncUserChangeUnit `json:"name"`
	Email       *UserSyncUserChangeUnit `json:"email"`
	SectionCode *UserSyncUserChangeUnit `json:"sectionCode"`
	JobTitle    *UserSyncUserChangeUnit `json:"jobTitle"`
}
