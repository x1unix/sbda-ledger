package ledger

type Group struct {
	ID      string `json:"id"`
	OwnerID string `json:"owner_id"`
	Name    string `json:"name"`
}

type GroupInfo struct {
	Group
	Members []User `json:"members"`
}

type groupCreateParams struct {
	Name string `json:"name"`
}

type groupsResponse struct {
	Groups []Group `json:"groups"`
}

type idsRequest struct {
	IDs []string `json:"ids"`
}

type amountRequest struct {
	Amount int64 `json:"amount"`
}

func (c Client) CreateGroup(name string, t Token) (*Group, error) {
	out := new(Group)
	return out, c.post("/groups", groupCreateParams{Name: name}, out, t)
}

func (c Client) Groups(t Token) ([]Group, error) {
	out := new(groupsResponse)
	return out.Groups, c.get("/groups", out, t)
}

func (c Client) GroupByID(gid string, t Token) (*GroupInfo, error) {
	out := new(GroupInfo)
	return out, c.get("/groups/"+gid, out, t)
}

func (c Client) DeleteGroup(gid string, t Token) error {
	return c.delete("/groups/"+gid, t)
}

func (c Client) GroupMembers(gid string, t Token) ([]User, error) {
	out := new(UsersResponse)
	return out.Users, c.get("/groups/"+gid+"/members", out, t)
}

func (c Client) AddGroupMembers(gid string, t Token, uids ...string) error {
	return c.post("/groups/"+gid+"/members", idsRequest{IDs: uids}, nil, t)
}

func (c Client) DeleteGroupMember(gid, uid string, t Token) error {
	return c.delete("/groups/"+gid+"/members/"+uid, t)
}

func (c Client) AddGroupExpense(gid string, amount int64, t Token) error {
	return c.post("/groups/"+gid+"/expenses", amountRequest{Amount: amount}, nil, t)
}
