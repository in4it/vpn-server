package login

import (
	"sync"
	"time"
)

var mu sync.Mutex

type Attempts map[string][]Attempt

type Attempt struct {
	Timestamp time.Time
}

func ClearAttemptsForLogin(attempts Attempts, login string) {
	mu.Lock()
	defer mu.Unlock()
	attempts[login] = []Attempt{}
}

func RecordAttempt(attempts Attempts, login string) {
	mu.Lock()
	defer mu.Unlock()
	_, ok := attempts[login]
	if !ok {
		attempts[login] = []Attempt{}
	}
	attempts[login] = append(attempts[login], Attempt{Timestamp: time.Now()})
}

func CheckTooManyLogins(attempts Attempts, login string) bool {
	threeMinutes := 3 * time.Minute
	_, ok := attempts[login]
	if ok {
		loginAttempts := 0
		for _, loginAttempt := range attempts[login] {
			if time.Since(loginAttempt.Timestamp) <= threeMinutes {
				loginAttempts++
			}
		}
		if loginAttempts >= 3 {
			if len(attempts[login]) > 3 {
				index := len(attempts[login]) - 3
				attempts[login] = attempts[login][index:]
			}
			return true
		}
	}
	return false
}
