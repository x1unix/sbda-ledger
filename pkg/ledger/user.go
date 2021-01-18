package ledger

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UsersResponse struct {
	Users []User `json:"users"`
}

type Balance struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}

type balanceResponse struct {
	Status []Balance `json:"status"`
}

func (c Client) Users(t Token) ([]User, error) {
	rsp := new(UsersResponse)
	return rsp.Users, c.get("/users", rsp, t)
}

func (c Client) UserByID(uid string, t Token) (*User, error) {
	rsp := new(User)
	return rsp, c.get("/users/"+uid, rsp, t)
}

func (c Client) CurrentUser(t Token) (*User, error) {
	rsp := new(User)
	return rsp, c.get("/users/self", rsp, t)
}

func (c Client) Balance(t Token) ([]Balance, error) {
	rsp := new(balanceResponse)
	return rsp.Status, c.get("/users/self/balance", rsp, t)
}
