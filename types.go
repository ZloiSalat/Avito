package main

type User struct {
	ID      int    `json:"id"`
	Segment string `json:"slug"`
}

type Request struct {
	UserID         int      `json:"user_id"`
	AddSegments    []string `json:"add_segments"`
	RemoveSegments []string `json:"remove_segments"`
}

func NewSegment(slug string) (*User, error) {
	return &User{
		Segment: slug,
	}, nil
}

func NewRequest(user_id int, add_segments []string, remove_segments []string) *Request {
	return &Request{
		UserID:         user_id,
		AddSegments:    add_segments,
		RemoveSegments: remove_segments,
	}
}
