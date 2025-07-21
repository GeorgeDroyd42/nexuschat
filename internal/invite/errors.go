package invite

const (
	ErrInvalidInviteCode = 1019
	ErrInviteExpired     = 1020
	ErrInviteNotFound    = 1021
)

var ErrorMessages = map[int]string{
	ErrInvalidInviteCode: "Invalid invite code",
	ErrInviteExpired:     "Invite code has expired", 
	ErrInviteNotFound:    "Invite code not found",
}