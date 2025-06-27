package main

type LoginResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateCommandRequest struct {
	FullName    string `json:"fullname"`
	Phone       string `json:"phone"`
	Service     string `json:"service"`
	Workers     string `json:"workers"`
	Start       string `json:"start"`
	Distination string `json:"distination"`
}

type Command struct {
	ID          int    `json:"id"`
	FullName    string `json:"firstName"`
	Number      string `json:"number"`
	Service     string `json:"service"`
	Workers     string `json:"workers"`
	Start       string `json:"start"`
	Distination string `json:"distination"`
}

func NewCommand(fullname, number, service, workers, start, distination string) (*Command, error) {
	return &Command{
		FullName:    fullname,
		Number:      number,
		Service:     service,
		Workers:     workers,
		Start:       start,
		Distination: distination,
	}, nil
}

type CreateWorkerRequest struct {
	FullName   string `json:"fullname"`
	Number     string `json:"number"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Position   string `json:"position"`
	Experience string `json:"experience"`
	Message    string `json:"message"`
	IsAccepted bool   `json:"isaccepted"`
}

type Worker struct {
	ID         string `json:"id"`
	FullName   string `json:"fullname"`
	Number     string `json:"number"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Position   string `json:"position"`
	Experience string `json:"experience"`
	Message    string `json:"message"`
	IsAccepted bool   `json:"isaccepted"`
}
