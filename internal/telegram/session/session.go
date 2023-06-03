package session

type Session struct {
	User   *User
	ChatID int64
	Spam   *spam
}

type User struct {
	Id    int64
	Login string
	Role  string
	state map[interface{}]interface{}
}

func (s *Session) Clear() {
	s.User = nil
}

func (s *Session) IsAuthenticated() bool {
	return s.User != nil
}

func MakeUser(id int64, login string, role string) *User {
	return &User{
		Id:    id,
		Login: login,
		Role:  role,
		state: make(map[interface{}]interface{}),
	}
}
