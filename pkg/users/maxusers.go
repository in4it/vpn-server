package users

func (u *UserStore) GetMaxUsers() int {
	return u.maxUsers
}
