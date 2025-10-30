package main

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	Key       string
	CreatedAt time.Time
}

type retentionMap map[string]OTP

func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) retentionMap {
	rm := make(retentionMap)

	go rm.Retention(ctx, retentionPeriod)

	return rm
}

func (rm retentionMap) NewOtp() OTP {
	o := OTP{
		Key:       uuid.NewString(),
		CreatedAt: time.Now(),
	}

	return o
}

func (rm retentionMap) VerifyOTP(otp string) bool {
	if _, ok := rm[otp]; !ok {
		return false
	}

	delete(rm, otp)
	return true
}

func (rm retentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			for _, otp := range rm {
				if otp.CreatedAt.Add(retentionPeriod).Before(time.Now()) {
					delete(rm, otp.Key)
					return
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
