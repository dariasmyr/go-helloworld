package requestlimiter

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	LimitSec int
}

type User struct {
	ID        string
	IsPremium bool
	TimeUsed  int
	Mx        *sync.Mutex
}

func (s *Service) HandleAllRequests(process func(), u *User) bool {
	done := make(chan bool)
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	go func() {
		process()
		done <- true
	}()

	fmt.Printf("Started handling requests for user: %s\n", u.ID)

	for {
		select {
		case <-done:
			fmt.Printf("Process func has been done for user: %s\n", u.ID)
			return true
		case <-ticker.C:
			u.Mx.Lock()
			u.TimeUsed++

			fmt.Printf("User %s: TimeUsed = %d / %d seconds\n", u.ID, u.TimeUsed, s.LimitSec)

			if !u.IsPremium && u.TimeUsed > s.LimitSec {
				u.Mx.Unlock()
				fmt.Printf("Time used for processing has reached the limit of %v sec \n", s.LimitSec)
				return false
			}

			u.Mx.Unlock()
		}
	}
}

// Context cancel should not allow to execute process() function correctly even for premium users
func (s *Service) HandleAllRequestsWithContext(process func(ctx context.Context), u *User) bool {
	done := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(s.LimitSec))
	defer cancel()

	go func() {
		process(ctx)
		close(done)
	}()

	fmt.Printf("Started handling requests for user: %s\n", u.ID)

	for {
		select {
		case <-done:
			fmt.Printf("Process func has been done for user: %s\n", u.ID)
			return true
		case <-ctx.Done():
			fmt.Printf("User %s, limit: %d seconds\n", u.ID, s.LimitSec)

			if !u.IsPremium {
				fmt.Printf("Time limit of %v seconds exceeded for user: %s, error: %v\n", s.LimitSec, u.ID, ctx.Err())
				return false
			}
		}
	}
}

func (s *Service) HandleRequest(process func(), u *User) bool {
	done := make(chan bool)

	go func() {
		process()
		done <- true
	}()

	success := true

	fmt.Printf("Started handling request for user: %s\n", u.ID)

	select {
	case <-done:
		fmt.Printf("Process func has been done for user: %s\n", u.ID)
		return success
	case <-time.After(time.Second * time.Duration(s.LimitSec)):
		if !u.IsPremium {
			fmt.Printf("Time used for processing has reached the limit of %v sec \n", s.LimitSec)
			return false
		}
	}

	return success
}

func sampleProcess(seconds int) {
	start := time.Now()

	time.Sleep(time.Duration(seconds) * time.Second)
	log.Printf("Process finished after: %v", time.Since(start))
}

func sampleProcessWithContext(ctx context.Context, seconds int) {
	start := time.Now()

	for i := 0; i < seconds; i++ {

		select {
		case <-time.After(1 * time.Second):
			log.Printf("Working... %d second(s)", i+1)
		case <-ctx.Done():
			log.Printf("Process cancelled after: %v, reason: %v", time.Since(start), ctx.Err())
			return
		}
	}

	log.Printf("Process finished successfully after: %v", time.Since(start))
}

func TestHandlingAllRequests(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		service := &Service{
			LimitSec: 10,
		}

		user := &User{
			ID:        uuid.NewString(),
			IsPremium: false,
			Mx:        &sync.Mutex{},
		}

		for i := 1; i <= 15; i++ {
			successful := service.HandleAllRequests(func() { sampleProcess(2) }, user)
			log.Printf("Finished short process %v with success: %v, user premium: %v \n", i, successful, user.IsPremium)
		}
	})
}

func TestHandlingAllRequestsWithContext(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		service := &Service{
			LimitSec: 3,
		}

		user := &User{
			ID:        uuid.NewString(),
			IsPremium: false,
		}

		t.Run("Process respects context timeout", func(t *testing.T) {
			successful := service.HandleAllRequestsWithContext(
				func(ctx context.Context) { sampleProcessWithContext(ctx, 5) },
				user,
			)
			log.Printf("Process finished with success: %v", successful)
		})

		t.Run("Process completes within timeout", func(t *testing.T) {
			successful := service.HandleAllRequestsWithContext(
				func(ctx context.Context) { sampleProcessWithContext(ctx, 2) },
				user,
			)
			log.Printf("Process finished with success: %v", successful)
		})
	})
}

func TestHandlingRequest(t *testing.T) {
	t.Run(t.Name(), func(t *testing.T) {
		service := &Service{
			LimitSec: 3,
		}

		user := &User{
			ID:        uuid.NewString(),
			IsPremium: false,
		}

		for i := 1; i <= 5; i++ {
			successful := service.HandleRequest(func() { sampleProcess(2) }, user)
			log.Printf("Finished short process %v with success: %v, user premium: %v \n", i, successful, user.IsPremium)

			nonsuccessfull := service.HandleRequest(func() { sampleProcess(4) }, user)
			log.Printf("Finished short process %v with success: %v, user premium: %v \n", i, nonsuccessfull, user.IsPremium)
		}
	})
}
