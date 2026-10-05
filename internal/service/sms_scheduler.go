package service

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// StartBirthdayCron runs BirthdayRunner on cfg cron expr in the given timezone.
// Returns a stop function (call on shutdown).
func StartBirthdayCron(ctx context.Context, runner *BirthdayRunner, cronExpr, tzName string) (stop func(), err error) {
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, err
	}

	c := cron.New(cron.WithLocation(loc))
	_, err = c.AddFunc(cronExpr, func() {
		runCtx, cancel := context.WithTimeout(context.Background(), BirthdayRunTimeout)
		defer cancel()
		if err := runner.Run(runCtx, time.Now().In(loc)); err != nil {
			log.Printf("birthday cron error: %v", err)
		}
	})
	if err != nil {
		return nil, err
	}

	c.Start()
	log.Printf("sms: birthday cron started expr=%q tz=%s", cronExpr, tzName)
	return func() {
		stopCtx := c.Stop()
		<-stopCtx.Done()
		log.Printf("sms: birthday cron stopped")
	}, nil
}
